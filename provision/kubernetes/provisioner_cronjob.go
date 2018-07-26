package kubernetes

import (
	"github.com/tsuru/tsuru/provision"
	//"github.com/tsuru/tsuru/provision/servicecommon"
)

var (
	_ provision.CronjobProvisioner = &kubernetesProvisioner{}
)

func (p *kubernetesProvisioner) GetCronjobs(appName string) ([]provision.CronJob, error) {
	return make([]provision.CronJob, 1), nil
}

func (p *kubernetesProvisioner) DeleteCronjob(appName, jobName string) error {
	return nil
}

func (p *kubernetesProvisioner) AddCronjob(appName string, jobSpec provision.CronJob) error {
	return nil
}
