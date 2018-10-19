package kubernetes

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	tsuruErrors "github.com/tsuru/tsuru/errors"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/dockercommon"
	"github.com/tsuru/tsuru/provision/servicecommon"
	batchv1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type cronjobManager struct {
	client *ClusterClient
	writer io.Writer
}

var _ servicecommon.CronjobManager = &cronjobManager{}

func (m *cronjobManager) RemoveCronjob(a provision.App, jobName string) error {
	ns, err := m.client.AppNamespace(a)
	if err != nil {
		return err
	}
	multiErrors := tsuruErrors.NewMultiError()
	cronjobName := cronJobNameForApp(a, jobName)
	err = m.client.BatchV1beta1().CronJobs(ns).Delete(cronjobName, &metav1.DeleteOptions{
		PropagationPolicy: propagationPtr(metav1.DeletePropagationForeground),
	})
	if err != nil && !k8sErrors.IsNotFound(err) {
		multiErrors.Add(errors.WithStack(err))
	}
	return multiErrors.ToError()
}

func (m *cronjobManager) CurrentLabels(a provision.App, processName string) (*provision.LabelSet, error) {
	return &provision.LabelSet{}, nil
}

func (m *cronjobManager) DeployCronjob(ctx context.Context, a provision.App, jobSpec *provision.CronJob, labels *provision.LabelSet, image string) error {
	err := ensureNodeContainers(a)
	if err != nil {
		return err
	}
	err = ensureNamespaceForApp(m.client, a)
	if err != nil {
		return err
	}
	err = ensureServiceAccountForApp(m.client, a)
	if err != nil {
		return err
	}
	cronjobName := cronJobNameForApp(a, jobSpec.Name)
	ns, err := m.client.AppNamespace(a)
	if err != nil {
		return err
	}
	oldCronjob, err := m.client.BatchV1beta1().CronJobs(ns).Get(cronjobName, metav1.GetOptions{})
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return errors.WithStack(err)
		}
		oldCronjob = nil
	}

	_, _, _, err = createAppCronjob(m.client, oldCronjob, a, jobSpec, image, labels)
	if err != nil {
		return err
	}

	return nil
}

func cronJobNameForApp(a provision.App, jobName string) string {
	name := strings.ToLower(kubeNameRegex.ReplaceAllString(a.GetName(), "-"))
	jobName = strings.ToLower(kubeNameRegex.ReplaceAllString(jobName, "-"))
	return fmt.Sprintf("%s-%s", name, jobName)
}

func createAppCronjob(client *ClusterClient, oldCronjob *v1beta1.CronJob, a provision.App, jobSpec *provision.CronJob, imageName string, labels *provision.LabelSet) (*v1beta1.CronJob, *provision.LabelSet, *provision.LabelSet, error) {
	provision.ExtendServiceLabels(labels, provision.ServiceLabelExtendedOpts{
		Provisioner: provisionerName,
		Prefix:      tsuruLabelPrefix,
	})

	cmds := getCommand(jobSpec.Command)

	appEnvs := provision.EnvsForApp(a, jobSpec.Name, false)
	var envs []apiv1.EnvVar
	for _, envData := range appEnvs {
		envs = append(envs, apiv1.EnvVar{Name: envData.Name, Value: envData.Value})
	}
	cronjobName := cronJobNameForApp(a, jobSpec.Name)

	nodeSelector := provision.NodeLabels(provision.NodeLabelsOpts{
		Pool:   a.GetPool(),
		Prefix: tsuruLabelPrefix,
	}).ToNodeByPoolSelector()
	_, uid := dockercommon.UserForContainer()
	resourceLimits := apiv1.ResourceList{}
	overcommit, err := client.OvercommitFactor(a.GetPool())
	if err != nil {
		return nil, nil, nil, errors.WithMessage(err, "misconfigured cluster overcommit factor")
	}
	resourceRequests := apiv1.ResourceList{}
	memory := a.GetMemory()
	if memory != 0 {
		resourceLimits[apiv1.ResourceMemory] = *resource.NewQuantity(memory, resource.BinarySI)
		resourceRequests[apiv1.ResourceMemory] = *resource.NewQuantity(memory/overcommit, resource.BinarySI)
	}
	volumes, mounts, err := createVolumesForApp(client, a)
	if err != nil {
		return nil, nil, nil, err
	}
	ns, err := client.AppNamespace(a)
	if err != nil {
		return nil, nil, nil, err
	}
	pullSecrets, err := getImagePullSecrets(client, ns, imageName)
	if err != nil {
		return nil, nil, nil, err
	}
	labels, annotations := provision.SplitServiceLabelsAnnotations(labels)
	failedJobs := int32(jobSpec.FailedJobsHistoryLimit)
	successJobs := int32(jobSpec.SuccessfulJobsHistoryLimit)
	cronjob := v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cronjobName,
			Namespace:   ns,
			Labels:      labels.WithoutAppReplicas().ToLabels(),
			Annotations: annotations.ToLabels(),
		},
		Spec: v1beta1.CronJobSpec{
			Schedule:                   jobSpec.Schedule,
			ConcurrencyPolicy:          v1beta1.ConcurrencyPolicy(jobSpec.ConcurrencyPolicy),
			FailedJobsHistoryLimit:     &failedJobs,
			SuccessfulJobsHistoryLimit: &successJobs,
			Suspend:                    &jobSpec.Suspend,
			JobTemplate: v1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels.WithoutAppReplicas().ToLabels(),
					Annotations: annotations.ToLabels(),
				},
				Spec: batchv1.JobSpec{
					Template: apiv1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:      labels.WithoutAppReplicas().ToLabels(),
							Annotations: annotations.ToLabels(),
						},
						Spec: apiv1.PodSpec{
							ImagePullSecrets:   pullSecrets,
							ServiceAccountName: serviceAccountNameForApp(a),
							SecurityContext: &apiv1.PodSecurityContext{
								RunAsUser: uid,
							},
							RestartPolicy: apiv1.RestartPolicyNever,
							NodeSelector:  nodeSelector,
							Volumes:       volumes,
							//Subdomain:     headlessServiceNameForApp(a, process),
							Containers: []apiv1.Container{
								{
									Name:    cronjobName,
									Image:   imageName,
									Command: cmds,
									Env:     envs,
									//ReadinessProbe: hcData.readiness,
									//LivenessProbe:  hcData.liveness,
									Resources: apiv1.ResourceRequirements{
										Limits:   resourceLimits,
										Requests: resourceRequests,
									},
									VolumeMounts: mounts,
									//Ports: []apiv1.ContainerPort{
									//	{ContainerPort: int32(portInt)},
									//},
									//Lifecycle: lifecycle,
								},
							},
						},
					},
				},
			},
		},
	}
	var newCronjob *v1beta1.CronJob
	if oldCronjob == nil {
		newCronjob, err = client.BatchV1beta1().CronJobs(ns).Create(&cronjob)
	} else {
		newCronjob, err = client.BatchV1beta1().CronJobs(ns).Update(&cronjob)
	}
	return newCronjob, labels, annotations, errors.WithStack(err)
}

func getCommand(command string) []string {
	allCmds := []string{
		"/bin/sh",
		"-lc",
		"[ -d /home/application/current ] && cd /home/application/current; exec " + command,
	}
	return allCmds
}
