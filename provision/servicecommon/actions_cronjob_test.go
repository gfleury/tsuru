package servicecommon

import (
	"context"
	//"testing"

	"github.com/pkg/errors"
	//"github.com/tsuru/config"
	"github.com/tsuru/tsuru/action"
	//"github.com/tsuru/tsuru/app"
	"github.com/tsuru/tsuru/app/image"
	//"github.com/tsuru/tsuru/db"
	//"github.com/tsuru/tsuru/db/dbtest"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/provisiontest"
	//servicemock "github.com/tsuru/tsuru/servicemanager/mock"
	//appTypes "github.com/tsuru/tsuru/types/app"
	"gopkg.in/check.v1"
)

type cronjobManagerCall struct {
	action  string
	app     provision.App
	jobSpec provision.CronJob
	image   string
	labels  *provision.LabelSet
}

func (m *recordManager) DeployCronjob(ctx context.Context, a provision.App, jobSpec *provision.CronJob, labels *provision.LabelSet, image string) error {
	call := cronjobManagerCall{
		action:  "deploy",
		jobSpec: *jobSpec,
		image:   image,
		labels:  labels,
		app:     a,
	}
	m.cronCalls = append(m.cronCalls, call)
	if m.deployErrMap != nil {
		return m.deployErrMap[jobSpec.Name]
	}
	return nil
}

func (m *recordManager) RemoveCronjob(a provision.App, jobName string) error {
	jobSpec := provision.CronJob{
		Name: jobName,
	}
	call := cronjobManagerCall{
		action:  "remove",
		jobSpec: jobSpec,
		app:     a,
	}
	m.cronCalls = append(m.cronCalls, call)
	if m.removeErrMap != nil {
		return m.removeErrMap[jobSpec.Name]
	}
	return nil
}

func (s *S) TestRunCronjobPipeline(c *check.C) {
	m := &recordManager{}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	err := image.SaveImageCustomData("oldImage", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "web",
			},
			1: map[string]string{
				"name": "worker1",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = image.AppendAppImageName(fakeApp.GetName(), "oldImage")
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("newImage", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "web",
			},
			1: map[string]string{
				"name": "worker2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = RunCronjobPipeline(m, fakeApp, "newImage", CronjobSpec{
		"web":     provision.CronJob{Name: "web"},
		"worker2": provision.CronJob{Name: "worker2"},
	}, nil)
	c.Assert(err, check.IsNil)
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	labelsWorker, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "worker2",
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "newImage", labels: labelsWeb},
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "worker2"}, image: "newImage", labels: labelsWorker},
		{action: "remove", app: fakeApp, jobSpec: provision.CronJob{Name: "worker1"}},
	})
	imgName, err := image.AppCurrentImageName(fakeApp.GetName())
	c.Assert(err, check.IsNil)
	c.Assert(imgName, check.Equals, "oldImage")
}

func (s *S) TestRunCronjobPipelineNilSpec(c *check.C) {
	m := &recordManager{}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	err := image.SaveImageCustomData("oldImage", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "web",
			},
			1: map[string]string{
				"name": "worker1",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = image.AppendAppImageName(fakeApp.GetName(), "oldImage")
	c.Assert(err, check.IsNil)
	err = image.SaveImageCustomData("newImage", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "web",
			},
			1: map[string]string{
				"name": "worker2",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = RunCronjobPipeline(m, fakeApp, "newImage", nil, nil)
	c.Assert(err, check.IsNil)
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	labelsWorker, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "worker2",
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "newImage", labels: labelsWeb},
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "worker2"}, image: "newImage", labels: labelsWorker},
		{action: "remove", app: fakeApp, jobSpec: provision.CronJob{Name: "worker1"}},
	})
	imgName, err := image.AppCurrentImageName(fakeApp.GetName())
	c.Assert(err, check.IsNil)
	c.Assert(imgName, check.Equals, "oldImage")
}

func (s *S) TestRunCronjobPipelineSingleProcess(c *check.C) {
	m := &recordManager{}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	err := image.SaveImageCustomData("oldImage", map[string]interface{}{
		"cronjobs": []interface{}{
			0: map[string]string{
				"name": "web",
			},
			1: map[string]string{
				"name": "worker1",
			},
		},
	})
	c.Assert(err, check.IsNil)
	err = image.AppendAppImageName(fakeApp.GetName(), "oldImage")
	c.Assert(err, check.IsNil)
	err = RunCronjobPipeline(m, fakeApp, "oldImage", CronjobSpec{
		"web":     provision.CronJob{Name: "web"},
		"worker1": provision.CronJob{Name: "worker1"},
	}, nil)
	c.Assert(err, check.IsNil)
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	labelsWorker, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "worker1",
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "oldImage", labels: labelsWeb},
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "worker1"}, image: "oldImage", labels: labelsWorker},
	})
}

func (s *S) TestActionUpdateCronjobForward(c *check.C) {
	m := &recordManager{}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	args := &pipelineCronjobsArgs{
		manager:          m,
		app:              fakeApp,
		newImage:         "image",
		newImageSpec:     CronjobSpec{"web": provision.CronJob{Name: "web"}},
		currentImage:     "oldImage",
		currentImageSpec: CronjobSpec{},
	}
	processes, err := updateCronjobs.Forward(action.FWContext{Params: []interface{}{args}})
	c.Assert(err, check.IsNil)
	c.Assert(processes, check.DeepEquals, []string{"web"})
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "image", labels: labelsWeb},
	})
	c.Assert(fakeApp.Quota.InUse, check.Equals, 0)
}

func (s *S) TestActionUpdateCronjobForwardMultiple(c *check.C) {
	m := &recordManager{}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	args := &pipelineCronjobsArgs{
		manager:          m,
		app:              fakeApp,
		newImage:         "image",
		newImageSpec:     CronjobSpec{"web": provision.CronJob{Name: "web"}, "worker2": provision.CronJob{Name: "worker2"}},
		currentImage:     "oldImage",
		currentImageSpec: CronjobSpec{"web": provision.CronJob{Name: "web"}, "worker1": provision.CronJob{Name: "worker1"}},
	}
	processes, err := updateCronjobs.Forward(action.FWContext{Params: []interface{}{args}})
	c.Assert(err, check.IsNil)
	c.Assert(processes, check.DeepEquals, []string{"web", "worker2"})
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	labelsWorker, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "worker2",
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "image", labels: labelsWeb},
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "worker2"}, image: "image", labels: labelsWorker},
	})
	c.Assert(fakeApp.Quota.InUse, check.Equals, 0)
}

func (s *S) TestActionUpdateCronjobForwardFailureInMiddle(c *check.C) {
	expectedError := errors.New("my deploy error")
	m := &recordManager{
		deployErrMap: map[string]error{"worker2": expectedError},
	}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	args := &pipelineCronjobsArgs{
		manager:          m,
		app:              fakeApp,
		newImage:         "image",
		newImageSpec:     CronjobSpec{"web": provision.CronJob{Name: "web"}, "worker2": provision.CronJob{Name: "worker2"}},
		currentImage:     "oldImage",
		currentImageSpec: CronjobSpec{"web": provision.CronJob{Name: "web"}, "worker1": provision.CronJob{Name: "worker1"}},
	}
	processes, err := updateCronjobs.Forward(action.FWContext{Params: []interface{}{args}})
	c.Assert(err, check.Equals, expectedError)
	c.Assert(processes, check.IsNil)
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	labelsWebOld, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	labelsWorker, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:      fakeApp,
		Process:  "worker2",
		Replicas: 0,
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "image", labels: labelsWeb},
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "worker2"}, image: "image", labels: labelsWorker},
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "oldImage", labels: labelsWebOld},
	})
}

func (s *S) TestActionUpdateCronjobForwardFailureInMiddleNewProc(c *check.C) {
	expectedError := errors.New("my deploy error")
	m := &recordManager{
		deployErrMap: map[string]error{"worker2": expectedError},
	}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	args := &pipelineCronjobsArgs{
		manager:          m,
		app:              fakeApp,
		newImage:         "image",
		newImageSpec:     CronjobSpec{"web": provision.CronJob{Name: "web"}, "worker2": provision.CronJob{Name: "worker2"}},
		currentImage:     "oldImage",
		currentImageSpec: CronjobSpec{"worker1": provision.CronJob{Name: "worker1"}},
	}
	processes, err := updateCronjobs.Forward(action.FWContext{Params: []interface{}{args}})
	c.Assert(err, check.Equals, expectedError)
	c.Assert(processes, check.IsNil)
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "web",
	})
	c.Assert(err, check.IsNil)
	labelsWorker, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     fakeApp,
		Process: "worker2",
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "image", labels: labelsWeb},
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "worker2"}, image: "image", labels: labelsWorker},
		{action: "remove", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}},
	})
}

func (s *S) TestActionUpdateCronjobBackward(c *check.C) {
	m := &recordManager{}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	args := &pipelineCronjobsArgs{
		manager:          m,
		app:              fakeApp,
		newImage:         "image",
		newImageSpec:     CronjobSpec{"web": provision.CronJob{Name: "web"}, "worker2": provision.CronJob{Name: "worker2"}},
		currentImage:     "oldImage",
		currentImageSpec: CronjobSpec{"web": provision.CronJob{Name: "web"}, "worker1": provision.CronJob{Name: "worker1"}},
	}
	updateCronjobs.Backward(action.BWContext{
		FWResult: []string{"web", "worker2"},
		Params:   []interface{}{args},
	})
	labelsWeb, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:      fakeApp,
		Process:  "web",
		Replicas: 0,
	})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "deploy", app: fakeApp, jobSpec: provision.CronJob{Name: "web"}, image: "oldImage", labels: labelsWeb},
		{action: "remove", app: fakeApp, jobSpec: provision.CronJob{Name: "worker2"}},
	})
}

func (s *S) TestRemoveOldCronjobForward(c *check.C) {
	m := &recordManager{}
	fakeApp := provisiontest.NewFakeApp("myapp", "whitespace", 1)
	args := &pipelineCronjobsArgs{
		manager:          m,
		app:              fakeApp,
		newImage:         "image",
		newImageSpec:     CronjobSpec{"cronjob-1": provision.CronJob{Name: "cronjob-1"}, "worker2": provision.CronJob{Name: "worker2"}},
		currentImage:     "oldImage",
		currentImageSpec: CronjobSpec{"worker1": provision.CronJob{Name: "worker1"}, "worker2": provision.CronJob{Name: "worker2"}},
	}
	_, err := removeOldCronjobs.Forward(action.FWContext{Params: []interface{}{args}})
	c.Assert(err, check.IsNil)
	c.Assert(m.cronCalls, check.DeepEquals, []cronjobManagerCall{
		{action: "remove", app: fakeApp, jobSpec: provision.CronJob{Name: "worker1"}},
	})
}
