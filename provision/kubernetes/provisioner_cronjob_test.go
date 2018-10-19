package kubernetes

import (
	"strings"

	"github.com/tsuru/tsuru/app"
	"github.com/tsuru/tsuru/provision"
	"gopkg.in/check.v1"
)

func (s *S) TestGetCronjobs(c *check.C) {
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	_, err = s.p.AddCronjob(a, provision.CronJob{
		Command:                    "ls -la",
		ConcurrencyPolicy:          "allow",
		FailedJobsHistoryLimit:     3,
		Name:                       "cr1",
		Schedule:                   "*/1 * * * *",
		SuccessfulJobsHistoryLimit: 3,
		Suspend:                    false,
	})
	c.Assert(err, check.IsNil)
	cronJobs, err := s.p.GetCronjobs(a)
	c.Assert(err, check.IsNil)
	c.Assert(cronJobs, check.DeepEquals, []provision.CronJob{{
		Name:                       "cr1",
		Command:                    strings.Join(getCommand("ls -la"), " "),
		ConcurrencyPolicy:          "allow",
		FailedJobsHistoryLimit:     3,
		Schedule:                   "*/1 * * * *",
		SuccessfulJobsHistoryLimit: 3,
		Suspend:                    false,
	}})
}

func (s *S) TestDeleteCronjob(c *check.C) {
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	_, err = s.p.AddCronjob(a, provision.CronJob{
		Command:                    "ls -la",
		ConcurrencyPolicy:          "allow",
		FailedJobsHistoryLimit:     3,
		Name:                       "cr1",
		Schedule:                   "*/1 * * * *",
		SuccessfulJobsHistoryLimit: 3,
		Suspend:                    false,
	})
	c.Assert(err, check.IsNil)
	err = s.p.DeleteCronjob(a, "cr1")
	c.Assert(err, check.IsNil)
	cronJobs, err := s.p.GetCronjobs(a)
	c.Assert(err, check.IsNil)
	c.Assert(cronJobs, check.DeepEquals, []provision.CronJob{})
}

func (s *S) TestAddCronjob(c *check.C) {
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	_, err = s.p.AddCronjob(a, provision.CronJob{
		Command:                    "ls -la",
		ConcurrencyPolicy:          "allow",
		FailedJobsHistoryLimit:     3,
		Name:                       "cr1",
		Schedule:                   "*/1 * * * *",
		SuccessfulJobsHistoryLimit: 3,
		Suspend:                    false,
	})
	c.Assert(err, check.IsNil)
	cronJobs, err := s.p.GetCronjobs(a)
	c.Assert(err, check.IsNil)
	c.Assert(cronJobs, check.DeepEquals, []provision.CronJob{{
		Name:                       "cr1",
		Command:                    strings.Join(getCommand("ls -la"), " "),
		ConcurrencyPolicy:          "allow",
		FailedJobsHistoryLimit:     3,
		Schedule:                   "*/1 * * * *",
		SuccessfulJobsHistoryLimit: 3,
		Suspend:                    false,
	}})
}

func (s *S) TestUpdateCronjob(c *check.C) {
	a := &app.App{Name: "myapp", TeamOwner: s.team.Name}
	err := app.CreateApp(a, s.user)
	c.Assert(err, check.IsNil)
	_, err = s.p.AddCronjob(a, provision.CronJob{
		Command:                    "ls -la",
		ConcurrencyPolicy:          "allow",
		FailedJobsHistoryLimit:     3,
		Name:                       "cr1",
		Schedule:                   "*/1 * * * *",
		SuccessfulJobsHistoryLimit: 3,
		Suspend:                    false,
	})
	c.Assert(err, check.IsNil)

	err = s.p.UpdateCronjob(a, "cr1", provision.CronJob{
		Command:           "cd /dist && ls -la /tmp",
		ConcurrencyPolicy: "allow",
		Schedule:          "1 10 10 * *",
		Suspend:           false,
	})
	c.Assert(err, check.IsNil)
	cronJobs, err := s.p.GetCronjobs(a)
	c.Assert(err, check.IsNil)
	c.Assert(cronJobs, check.DeepEquals, []provision.CronJob{{
		Name:                       "cr1",
		Command:                    strings.Join(getCommand("cd /dist && ls -la /tmp"), " "),
		ConcurrencyPolicy:          "allow",
		FailedJobsHistoryLimit:     3,
		Schedule:                   "1 10 10 * *",
		SuccessfulJobsHistoryLimit: 3,
		Suspend:                    false,
	}})
}
