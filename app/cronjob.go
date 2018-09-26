package app

import (
	"github.com/tsuru/tsuru/provision"
)

func (app *App) AddCronjob(cronjob *provision.CronJob) (string, error) {

	return "", nil
}

func (app *App) ListCronjobs() ([]provision.CronJob, error) {

	return []provision.CronJob{}, nil
}

func (app *App) DeleteCronjobs(name string) error {

	return nil
}

func (app *App) UpdateCronjob(cron provision.CronJob) (string, error) {

	return cron.Name, nil
}
