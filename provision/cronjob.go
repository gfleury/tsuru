package provision

type CronjobProvisioner interface {
	GetCronjobs(a App) ([]CronJob, error)
	DeleteCronjob(a App, jobName string) error
	AddCronjob(a App, jobSpec CronJob) (string, error)
	UpdateCronjob(a App, jobName string, cronJob CronJob) error
}

type CronJob struct {
	Name                       string `json:"name" yaml:"name" bson:"name,omitempty"`
	Schedule                   string `json:"schedule" yaml:"schedule" bson:"schedule,omitempty"`
	ConcurrencyPolicy          string `json:"concurrencyPolicy" yaml:"concurrencyPolicy" bson:"concurrencyPolicy,omitempty"`
	Command                    string `json:"command" yaml:"command" bson:"command,omitempty"`
	Suspend                    bool   `json:"suspend" yaml:"suspend" bson:"suspend,omitempty"`
	SuccessfulJobsHistoryLimit int    `json:"successfulJobsHistoryLimit" yaml:"successfulJobsHistoryLimit" bson:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     int    `json:"failedJobsHistoryLimit" yaml:"failedJobsHistoryLimit" bson:"failedJobsHistoryLimit,omitempty"`
}
