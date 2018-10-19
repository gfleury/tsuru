package app

import (
	"github.com/pkg/errors"
	"github.com/tsuru/tsuru/provision"
)

func (app *App) AddCronjob(cronjob *provision.CronJob) (string, error) {
	prov, err := app.getProvisioner()
	if err != nil {
		return "", err
	}
	provisioner, ok := prov.(provision.CronjobProvisioner)
	if !ok {
		return "", errors.Errorf("provisioner don't implement cronjob interface")
	}
	cronjobName, err := provisioner.AddCronjob(app, *cronjob)
	if err != nil {
		return "", err
	}
	return cronjobName, nil
}

func (app *App) ListCronjobs() ([]provision.CronJob, error) {
	prov, err := app.getProvisioner()
	if err != nil {
		return []provision.CronJob{}, err
	}
	provisioner, ok := prov.(provision.CronjobProvisioner)
	if !ok {
		return []provision.CronJob{}, errors.Errorf("provisioner don't implement cronjob interface")
	}
	cronjobs, err := provisioner.GetCronjobs(app)
	return cronjobs, err

}

func (app *App) DeleteCronjobs(name string) error {
	prov, err := app.getProvisioner()
	if err != nil {
		return err
	}
	provisioner, ok := prov.(provision.CronjobProvisioner)
	if !ok {
		return errors.Errorf("provisioner don't implement cronjob interface")
	}
	err = provisioner.DeleteCronjob(app, name)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) UpdateCronjob(cron provision.CronJob) (string, error) {
	prov, err := app.getProvisioner()
	if err != nil {
		return "", err
	}
	provisioner, ok := prov.(provision.CronjobProvisioner)
	if !ok {
		return "", errors.Errorf("provisioner don't implement cronjob interface")
	}
	err = provisioner.UpdateCronjob(app, cron.Name, cron)
	if err != nil {
		return "", err
	}
	return cron.Name, nil
}
