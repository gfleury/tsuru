package app

import (
	"github.com/tsuru/tsuru/provision"
	"gopkg.in/check.v1"
)

func (s *S) TestAddCronjob(c *check.C) {
	a := App{
		Name:      "some-app",
		Platform:  "django",
		Teams:     []string{s.team.Name},
		TeamOwner: s.team.Name,
		Router:    "fake",
	}
	err := CreateApp(&a, s.user)
	c.Assert(err, check.IsNil)
	cronjob := provision.CronJob{
		Command:                    "ls -la",
		ConcurrencyPolicy:          "Allow",
		FailedJobsHistoryLimit:     3,
		SuccessfulJobsHistoryLimit: 3,
		Name:                       "Fakecron",
		Schedule:                   "1/* * * * *",
		Suspend:                    false,
	}
	jobName, err := a.AddCronjob(&cronjob)
	c.Assert(err, check.IsNil)
	c.Assert(jobName, check.Equals, "some-app-fakecron")
}

func (s *S) TestUpdateCronjob(c *check.C) {
	a := App{
		Name:      "some-app",
		Platform:  "django",
		Teams:     []string{s.team.Name},
		TeamOwner: s.team.Name,
		Router:    "fake",
	}
	err := CreateApp(&a, s.user)
	c.Assert(err, check.IsNil)
	cronjob := provision.CronJob{
		Command:                    "ls -la",
		ConcurrencyPolicy:          "Allow",
		FailedJobsHistoryLimit:     3,
		SuccessfulJobsHistoryLimit: 3,
		Name:                       "Fakecron",
		Schedule:                   "1/* * * * *",
		Suspend:                    false,
	}
	jobName, err := a.UpdateCronjob(cronjob)
	c.Assert(err, check.IsNil)
	c.Assert(jobName, check.Equals, "Fakecron")
}

func (s *S) TestDeleteCronjob(c *check.C) {
	a := App{
		Name:      "some-app",
		Platform:  "django",
		Teams:     []string{s.team.Name},
		TeamOwner: s.team.Name,
		Router:    "fake",
	}
	err := CreateApp(&a, s.user)
	c.Assert(err, check.IsNil)
	err = a.DeleteCronjobs("some-app-fakecron")
	c.Assert(err, check.IsNil)
}
