package servicecommon

import (
	"context"
	"regexp"
	"sort"

	//"github.com/pkg/errors"
	"github.com/tsuru/tsuru/action"
	"github.com/tsuru/tsuru/app/image"
	"github.com/tsuru/tsuru/event"
	"github.com/tsuru/tsuru/log"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/set"
)

type CronjobSpec map[string]provision.CronJob

type pipelineCronjobsArgs struct {
	manager          CronjobManager
	app              provision.App
	newImage         string
	newImageSpec     CronjobSpec
	currentImage     string
	currentImageSpec CronjobSpec
	event            *event.Event
}

type CronjobManager interface {
	RemoveCronjob(a provision.App, jobName string) error
	CurrentLabels(a provision.App, jobName string) (*provision.LabelSet, error)
	DeployCronjob(ctx context.Context, a provision.App, jobSpec *provision.CronJob, labels *provision.LabelSet, image string) error
}

func RunCronjobPipeline(manager CronjobManager, a provision.App, newImg string, updateSpec CronjobSpec, evt *event.Event) error {
	curImg, err := image.AppPreviousImageName(a.GetName())
	if err != nil {
		return err
	}
	currentImageData, err := image.GetImageTsuruYamlData(curImg)
	if err != nil {
		return err
	}
	currentSpec := CronjobSpec{}
	for _, p := range currentImageData.Cronjobs {
		currentSpec[p.Name] = p
	}
	newImageData, err := image.GetImageTsuruYamlData(newImg)
	if err != nil {
		return err
	}

	newSpec := CronjobSpec{}
	for _, p := range newImageData.Cronjobs {
		newSpec[p.Name] = p
		if updateSpec != nil {
			newSpec[p.Name] = updateSpec[p.Name]
		}
	}
	pipeline := action.NewPipeline(
		updateCronjobs,
		removeOldCronjobs,
	)
	return pipeline.Execute(&pipelineCronjobsArgs{
		manager:          manager,
		app:              a,
		newImage:         newImg,
		newImageSpec:     newSpec,
		currentImage:     curImg,
		currentImageSpec: currentSpec,
		event:            evt,
	})
}

func rollbackAddedCronjobs(args *pipelineCronjobsArgs, cronjobs []string) {
	for _, cronjobName := range cronjobs {
		var err error
		if state, in := args.currentImageSpec[cronjobName]; in {
			var labels *provision.LabelSet
			labels, err = labelsForCronjobs(args, cronjobName, state)
			if err == nil {
				err = args.manager.DeployCronjob(context.Background(), args.app, &state, labels, args.currentImage)
			}
		} else {
			err = args.manager.RemoveCronjob(args.app, cronjobName)
		}
		if err != nil {
			log.Errorf("error rolling back updated service for %s[%s]: %+v", args.app.GetName(), cronjobName, err)
		}
	}
}

var kubeNameRegex = regexp.MustCompile(`(?i)[^a-z0-9.-]`)

func labelsForCronjobs(args *pipelineCronjobsArgs, cronjobName string, pState provision.CronJob) (*provision.LabelSet, error) {
	//oldLabels, err := args.manager.CurrentLabels(args.app, cronjobName)
	//if err != nil {
	//	return nil, err
	//}
	labels, err := provision.ServiceLabels(provision.ServiceLabelsOpts{
		App:     args.app,
		Process: kubeNameRegex.ReplaceAllString(cronjobName, "-"),
	})
	if err != nil {
		return nil, err
	}
	return labels, nil
}

var updateCronjobs = &action.Action{
	Name: "update-cronjobs",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(*pipelineCronjobsArgs)
		var (
			toDeployCronjobs []string
			deployedCronjobs []string
			err              error
		)
		for processName := range args.newImageSpec {
			toDeployCronjobs = append(toDeployCronjobs, processName)
		}
		sort.Strings(toDeployCronjobs)
		labelsMap := map[string]*provision.LabelSet{}
		for _, processName := range toDeployCronjobs {
			var labels *provision.LabelSet
			labels, err = labelsForCronjobs(args, processName, args.newImageSpec[processName])
			if err != nil {
				return nil, err
			}
			labelsMap[processName] = labels
		}

		if err != nil {
			return nil, err
		}
		for _, processName := range toDeployCronjobs {
			labels := labelsMap[processName]
			ectx, cancel := args.event.CancelableContext(context.Background())
			cronJob := args.newImageSpec[processName]
			err = args.manager.DeployCronjob(ectx, args.app, &cronJob, labels, args.newImage)
			cancel()
			if err != nil {
				break
			}
			deployedCronjobs = append(deployedCronjobs, processName)
		}
		if err != nil {
			rollbackAddedCronjobs(args, deployedCronjobs)
			return nil, err
		}
		return deployedCronjobs, nil
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(*pipelineCronjobsArgs)
		deployedProcesses := ctx.FWResult.([]string)
		rollbackAddedCronjobs(args, deployedProcesses)
	},
}

var removeOldCronjobs = &action.Action{
	Name: "remove-old-cronjobs",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(*pipelineCronjobsArgs)
		old := set.FromMap(args.currentImageSpec)
		new := set.FromMap(args.newImageSpec)
		for processName := range old.Difference(new) {
			err := args.manager.RemoveCronjob(args.app, processName)
			if err != nil {
				log.Errorf("ignored error removing unwanted service for %s[%s]: %+v", args.app.GetName(), processName, err)
			}
		}
		return nil, nil
	},
}
