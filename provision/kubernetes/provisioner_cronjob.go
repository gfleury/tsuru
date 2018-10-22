package kubernetes

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/tsuru/tsuru/app/image"
	"github.com/tsuru/tsuru/provision"
	v2alpha1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	//"github.com/tsuru/tsuru/provision/servicecommon"
)

var (
	_ provision.CronjobProvisioner = &kubernetesProvisioner{}
)

func (p *kubernetesProvisioner) GetCronjobs(a provision.App) ([]provision.CronJob, error) {
	client, err := clusterForPool(a.GetPool())
	if err != nil {
		return nil, err
	}

	ls, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App: a,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	provision.ExtendServiceLabels(ls, provision.ServiceLabelExtendedOpts{
		Provisioner: provisionerName,
		Prefix:      tsuruLabelPrefix,
	})

	ns, err := client.AppNamespace(a)
	if err != nil {
		return nil, err
	}
	cronJobs, err := client.BatchV1beta1().CronJobs(ns).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set(ls.WithoutAppReplicas().ToSelector())).String(),
	})
	if err != nil {
		return nil, err
	}

	tsuruCronJobs := make([]provision.CronJob, len(cronJobs.Items))
	for i, cronJob := range cronJobs.Items {
		tsuruCronJobs[i].Command = strings.Join(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command, " ")
		tsuruCronJobs[i].ConcurrencyPolicy = string(cronJob.Spec.ConcurrencyPolicy)
		tsuruCronJobs[i].FailedJobsHistoryLimit = int(*cronJob.Spec.FailedJobsHistoryLimit)
		tsuruCronJobs[i].Name = cronJob.Name[len(a.GetName())+1:]
		tsuruCronJobs[i].Schedule = cronJob.Spec.Schedule
		if cronJob.Spec.SuccessfulJobsHistoryLimit != nil {
			tsuruCronJobs[i].SuccessfulJobsHistoryLimit = int(*cronJob.Spec.SuccessfulJobsHistoryLimit)
		}
		if cronJob.Spec.Suspend != nil {
			tsuruCronJobs[i].Suspend = *cronJob.Spec.Suspend
		}
	}

	return tsuruCronJobs, nil
}

func (p *kubernetesProvisioner) DeleteCronjob(a provision.App, jobName string) error {
	client, err := clusterForPool(a.GetPool())
	if err != nil {
		return err
	}
	ns, err := client.AppNamespace(a)
	if err != nil {
		return err
	}
	return client.BatchV1beta1().CronJobs(ns).Delete(cronJobNameForApp(a, jobName), &metav1.DeleteOptions{
		PropagationPolicy: propagationPtr(metav1.DeletePropagationForeground),
	})
}

func (p *kubernetesProvisioner) AddCronjob(a provision.App, jobSpec provision.CronJob) (string, error) {
	err := ensureNodeContainers(a)
	if err != nil {
		return "", err
	}
	client, err := clusterForPool(a.GetPool())
	if err != nil {
		return "", err
	}
	err = ensureNamespaceForApp(client, a)
	if err != nil {
		return "", err
	}
	err = ensureServiceAccountForApp(client, a)
	if err != nil {
		return "", err
	}

	curImg, err := image.AppCurrentImageName(a.GetName())
	if err != nil {
		return "", err
	}

	ls, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App: a,
	})
	if err != nil {
		return "", err
	}

	newCronjob, _, _, err := createAppCronjob(client, nil, a, &jobSpec, curImg, ls.WithoutAppReplicas())
	if err != nil {
		return "", err
	}

	return newCronjob.GetName(), err
}

func (p *kubernetesProvisioner) UpdateCronjob(a provision.App, jobName string, cronJob provision.CronJob) error {
	client, err := clusterForPool(a.GetPool())
	if err != nil {
		return err
	}
	ns, err := client.AppNamespace(a)
	if err != nil {
		return err
	}
	oldCronJob, err := client.BatchV1beta1().CronJobs(ns).Get(cronJobNameForApp(a, jobName), metav1.GetOptions{})
	if err != nil {
		return err
	}
	if len(cronJob.Command) > 0 {
		oldCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command = getCommand(cronJob.Command)
	}
	if len(cronJob.ConcurrencyPolicy) > 0 {
		oldCronJob.Spec.ConcurrencyPolicy = v2alpha1.ConcurrencyPolicy(cronJob.ConcurrencyPolicy)
	}
	if cronJob.FailedJobsHistoryLimit > 0 {
		failedJobsHistoryLimit := int32(cronJob.FailedJobsHistoryLimit)
		oldCronJob.Spec.FailedJobsHistoryLimit = &failedJobsHistoryLimit
	}
	if len(cronJob.Schedule) > 0 {
		oldCronJob.Spec.Schedule = cronJob.Schedule
	}
	if cronJob.SuccessfulJobsHistoryLimit > 0 {
		successfulJobsHistoryLimit := int32(cronJob.SuccessfulJobsHistoryLimit)
		oldCronJob.Spec.SuccessfulJobsHistoryLimit = &successfulJobsHistoryLimit
	}

	oldCronJob.Spec.Suspend = &cronJob.Suspend

	_, err = client.BatchV1beta1().CronJobs(ns).Update(oldCronJob)

	return err
}
