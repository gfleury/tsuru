package kubernetes

import (
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/tsuru/config"
	"github.com/tsuru/tsuru/app"
	"github.com/tsuru/tsuru/app/image"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/servicecommon"
	appTypes "github.com/tsuru/tsuru/types/app"
	"github.com/tsuru/tsuru/volume"
	"gopkg.in/check.v1"
	batchv1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func (s *S) TestCronjobManagerDeployCronjob(c *check.C) {
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", servicecommon.CronjobSpec{
		"p1": provision.CronJob{Name: "p1"},
	}, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).Get("myapp-p1", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	depLabels := map[string]string{
		"tsuru.io/is-tsuru":        "true",
		"tsuru.io/is-service":      "true",
		"tsuru.io/is-build":        "false",
		"tsuru.io/is-stopped":      "false",
		"tsuru.io/is-deploy":       "false",
		"tsuru.io/is-isolated-run": "false",
		"tsuru.io/app-name":        "myapp",
		"tsuru.io/app-process":     "p1",
		"tsuru.io/app-platform":    "",
		"tsuru.io/app-pool":        "test-default",
		"tsuru.io/provisioner":     "kubernetes",
		"tsuru.io/builder":         "",
	}
	podLabels := make(map[string]string)
	for k, v := range depLabels {
		if k == "tsuru.io/app-process-replicas" {
			continue
		}
		podLabels[k] = v
	}
	annotations := map[string]string{
		"tsuru.io/router-type": "fake",
		"tsuru.io/router-name": "fake",
	}
	nsName, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep.Spec.SuccessfulJobsHistoryLimit = nil
	dep.Spec.FailedJobsHistoryLimit = nil
	dep.Spec.JobTemplate.Spec.Template.Spec.SecurityContext = nil
	falseBool := false
	c.Assert(dep, check.DeepEquals, &v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "myapp-p1",
			Namespace:   "default",
			Labels:      depLabels,
			Annotations: annotations,
		},
		Spec: v1beta1.CronJobSpec{
			Suspend: &falseBool,
			JobTemplate: v1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      depLabels,
					Annotations: annotations,
				},
				Spec: batchv1.JobSpec{
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Volumes:        []apiv1.Volume(nil),
							InitContainers: []apiv1.Container(nil),
							Containers: []apiv1.Container{
								{
									Name:  "myapp-p1",
									Image: "myimg",
									Command: []string{
										"/bin/sh",
										"-lc",
										"[ -d /home/application/current ] && cd /home/application/current; exec ",
									},
									Args:       []string(nil),
									WorkingDir: "",
									Ports:      []apiv1.ContainerPort(nil),
									EnvFrom:    []apiv1.EnvFromSource(nil),
									Env: []apiv1.EnvVar{
										{Name: "TSURU_SERVICES", Value: "{}"},
										{Name: "TSURU_PROCESSNAME", Value: "p1"},
										{Name: "TSURU_HOST", Value: ""},
										{Name: "port", Value: "8888"},
										{Name: "PORT", Value: "8888"},
									},
									Resources: apiv1.ResourceRequirements{
										Limits: apiv1.ResourceList{}, Requests: apiv1.ResourceList{},
									},
									VolumeMounts:   []apiv1.VolumeMount(nil),
									VolumeDevices:  []apiv1.VolumeDevice(nil),
									LivenessProbe:  (*apiv1.Probe)(nil),
									ReadinessProbe: (*apiv1.Probe)(nil),
									Lifecycle:      (*apiv1.Lifecycle)(nil),
									Stdin:          false,
									StdinOnce:      false,
									TTY:            false,
								},
							},
							RestartPolicy:      "OnFailure",
							NodeSelector:       map[string]string{"tsuru.io/pool": "test-default"},
							ServiceAccountName: "app-myapp",
							ImagePullSecrets:   []apiv1.LocalObjectReference(nil),
						},
					},
				},
			},
		},
	})

	account, err := s.client.CoreV1().ServiceAccounts(nsName).Get("app-myapp", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(account, check.DeepEquals, &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-myapp",
			Namespace: nsName,
			Labels: map[string]string{
				"tsuru.io/is-tsuru":    "true",
				"tsuru.io/app-name":    "myapp",
				"tsuru.io/provisioner": "kubernetes",
			},
		},
	})
}

func (s *S) TestCronjobManagerDeployCronjobWithPoolNamespaces(c *check.C) {
	config.Set("kubernetes:use-pool-namespaces", true)
	defer config.Unset("kubernetes:use-pool-namespaces")
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	var counter int32
	s.client.PrependReactor("create", "namespaces", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		new := atomic.AddInt32(&counter, 1)
		ns, ok := action.(ktesting.CreateAction).GetObject().(*apiv1.Namespace)
		c.Assert(ok, check.Equals, true)
		if new == 2 {
			c.Assert(ns.ObjectMeta.Name, check.Equals, "tsuru-test-default")
		} else if new < 2 {
			c.Assert(ns.ObjectMeta.Name, check.Equals, s.client.Namespace())
		}
		return false, nil, nil
	})
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)

	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", servicecommon.CronjobSpec{
		"p1": provision.CronJob{Name: "p1"},
	}, nil)
	c.Assert(err, check.IsNil)
	c.Assert(atomic.LoadInt32(&counter), check.Equals, int32(3))
}

func (s *S) TestCronjobManagerDeployCronjobWithRegistryAuth(c *check.C) {
	config.Set("docker:registry", "myreg.com")
	config.Set("docker:registry-auth:username", "user")
	config.Set("docker:registry-auth:password", "pass")
	defer config.Unset("docker:registry")
	defer config.Unset("docker:registry-auth")
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("myreg.com/myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "web",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myreg.com/myimg", servicecommon.CronjobSpec{
		"web": provision.CronJob{Name: "web"},
	}, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).Get("myapp-web", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(dep.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets, check.DeepEquals, []apiv1.LocalObjectReference{
		{Name: "registry-myreg.com"},
	})
	secrets, err := s.client.CoreV1().Secrets(ns).List(metav1.ListOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(secrets.Items, check.DeepEquals, []apiv1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "registry-myreg.com",
				Namespace: "default",
			},
			Data: map[string][]byte{
				".dockerconfigjson": []byte(`{"auths":{"myreg.com":{"username":"user","password":"pass","auth":"dXNlcjpwYXNz"}}}`),
			},
			Type: "kubernetes.io/dockerconfigjson",
		},
	})
}

func (s *S) TestCronjobManagerDeployCronjobWithUID(c *check.C) {
	config.Set("docker:uid", 1001)
	defer config.Unset("docker:uid")
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", servicecommon.CronjobSpec{
		"p1": provision.CronJob{Name: "p1"},
	}, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).Get("myapp-p1", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	expectedUID := int64(1001)
	c.Assert(dep.Spec.JobTemplate.Spec.Template.Spec.SecurityContext, check.DeepEquals, &apiv1.PodSecurityContext{
		RunAsUser: &expectedUID,
	})
}

func (s *S) TestCronjobManagerDeployCronjobWithResourceRequirements(c *check.C) {
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	a.Plan = appTypes.Plan{Memory: 1024}
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", servicecommon.CronjobSpec{
		"p1": provision.CronJob{Name: "p1"},
	}, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).Get("myapp-p1", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	expectedMemory := resource.NewQuantity(1024, resource.BinarySI)
	c.Assert(dep.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources, check.DeepEquals, apiv1.ResourceRequirements{
		Limits: apiv1.ResourceList{
			apiv1.ResourceMemory: *expectedMemory,
		},
		Requests: apiv1.ResourceList{
			apiv1.ResourceMemory: *expectedMemory,
		},
	})
}

func (s *S) TestCronjobManagerDeployCronjobWithClusterWideOvercommitFactor(c *check.C) {
	s.clusterClient.CustomData[overcommitClusterKey] = "3"
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	a.Plan = appTypes.Plan{Memory: 1024}
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", servicecommon.CronjobSpec{
		"p1": provision.CronJob{Name: "p1"},
	}, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).Get("myapp-p1", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	expectedMemory := resource.NewQuantity(1024, resource.BinarySI)
	expectedMemoryRequest := resource.NewQuantity(341, resource.BinarySI)
	c.Assert(dep.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources, check.DeepEquals, apiv1.ResourceRequirements{
		Limits: apiv1.ResourceList{
			apiv1.ResourceMemory: *expectedMemory,
		},
		Requests: apiv1.ResourceList{
			apiv1.ResourceMemory: *expectedMemoryRequest,
		},
	})
}

func (s *S) TestCronjobManagerDeployCronjobWithClusterPoolOvercommitFactor(c *check.C) {
	s.clusterClient.CustomData[overcommitClusterKey] = "3"
	s.clusterClient.CustomData["test-default:"+overcommitClusterKey] = "2"
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	a.Plan = appTypes.Plan{Memory: 1024}
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", servicecommon.CronjobSpec{
		"p1": provision.CronJob{Name: "p1"},
	}, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).Get("myapp-p1", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	expectedMemory := resource.NewQuantity(1024, resource.BinarySI)
	expectedMemoryRequest := resource.NewQuantity(512, resource.BinarySI)
	c.Assert(dep.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources, check.DeepEquals, apiv1.ResourceRequirements{
		Limits: apiv1.ResourceList{
			apiv1.ResourceMemory: *expectedMemory,
		},
		Requests: apiv1.ResourceList{
			apiv1.ResourceMemory: *expectedMemoryRequest,
		},
	})
}

func (s *S) TestCronjobManagerDeployCronjobWithVolumes(c *check.C) {
	config.Set("docker:uid", 1001)
	defer config.Unset("docker:uid")
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
			1: map[string]string{
				"name": "p2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	config.Set("volume-plans:p1:kubernetes:plugin", "nfs")
	defer config.Unset("volume-plans")
	v := volume.Volume{
		Name: "v1",
		Opts: map[string]string{
			"path":         "/exports",
			"server":       "192.168.1.1",
			"capacity":     "20Gi",
			"access-modes": string(apiv1.ReadWriteMany),
		},
		Plan:      volume.VolumePlan{Name: "p1"},
		Pool:      "test-default",
		TeamOwner: "admin",
	}
	err = v.Create()
	c.Assert(err, check.IsNil)
	err = v.BindApp(a.GetName(), "/mnt", false)
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", servicecommon.CronjobSpec{
		"p1": provision.CronJob{Name: "p1"},
	}, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	dep, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).Get("myapp-p1", metav1.GetOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(dep.Spec.JobTemplate.Spec.Template.Spec.Volumes, check.DeepEquals, []apiv1.Volume{
		{
			Name: "v1-tsuru",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "v1-tsuru-claim",
					ReadOnly:  false,
				},
			},
		},
	})
	c.Assert(dep.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts, check.DeepEquals, []apiv1.VolumeMount{
		{
			Name:      "v1-tsuru",
			MountPath: "/mnt",
			ReadOnly:  false,
		},
	})
}

func (s *S) TestCronjobManagerRemoveService(c *check.C) {
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", nil, nil)
	c.Assert(err, check.IsNil)
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	err = m.RemoveCronjob(a, "p1")
	c.Assert(err, check.IsNil)
	deps, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).List(metav1.ListOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(deps.Items, check.HasLen, 0)
	srvs, err := s.client.CoreV1().Services(ns).List(metav1.ListOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(srvs.Items, check.HasLen, 0)
	pods, err := s.client.CoreV1().Pods(ns).List(metav1.ListOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(pods.Items, check.HasLen, 0)
	replicas, err := s.client.Clientset.AppsV1beta2().ReplicaSets(ns).List(metav1.ListOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(replicas.Items, check.HasLen, 0)
}

func (s *S) TestCronjobManagerRemoveServiceMiddleFailure(c *check.C) {
	m := cronjobManager{client: s.clusterClient}
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("myimg", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "p1",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = servicecommon.RunCronjobPipeline(&m, a, "myimg", nil, nil)
	c.Assert(err, check.IsNil)
	s.client.PrependReactor("delete", "cronjobs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("my dep err")
	})
	err = m.RemoveCronjob(a, "p1")
	c.Assert(err, check.ErrorMatches, "(?s).*my dep err.*")
	ns, err := s.client.AppNamespace(a)
	c.Assert(err, check.IsNil)
	deps, err := s.client.Clientset.BatchV1beta1().CronJobs(ns).List(metav1.ListOptions{})
	c.Assert(err, check.IsNil)
	c.Assert(deps.Items, check.HasLen, 1)
}
