package api

import (
	"encoding/json"
	"net/http"

	"github.com/ajg/form"
	"github.com/tsuru/tsuru/auth"
	"github.com/tsuru/tsuru/errors"
	tsuruErrors "github.com/tsuru/tsuru/errors"
	"github.com/tsuru/tsuru/event"
	"github.com/tsuru/tsuru/permission"
	"github.com/tsuru/tsuru/provision"
)

// title: add cronjob
// path: /apps/{appname}/cronjobs
// method: POST
// consume: application/x-www-form-urlencoded
// responses:
//   200: OK
//   400: Invalid data
//   403: Forbidden
//   404: Not found
func addCronjob(w http.ResponseWriter, r *http.Request, t auth.Token) (err error) {
	err = r.ParseForm()
	if err != nil {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: err.Error()}
	}
	var cronjob provision.CronJob
	dec := form.NewDecoder(nil)
	dec.IgnoreCase(true)
	dec.IgnoreUnknownKeys(true)
	dec.DecodeValues(&cronjob, r.Form)

	appName := r.URL.Query().Get(":appname")
	a, err := getAppFromContext(appName, r)
	if err != nil {
		return err
	}

	canBuild := permission.Check(t, permission.PermAppUpdateCronjobAdd, contextsForApp(&a)...)
	if !canBuild {
		return &tsuruErrors.HTTP{Code: http.StatusForbidden, Message: "User does not have permission to do this action in this app"}
	}

	evt, err := event.New(&event.Opts{
		Target:     appTarget(appName),
		Kind:       permission.PermAppUpdateCronjobAdd,
		Owner:      t,
		CustomData: event.FormToCustomData(r.Form),
		Allowed:    event.Allowed(permission.PermAppReadEvents, contextsForApp(&a)...),
	})
	if err != nil {
		return err
	}
	defer func() { evt.Done(err) }()

	name, err := a.AddCronjob(&cronjob)
	if err != nil {
		return err
	}

	msg := map[string]string{
		"cronjob": name,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonMsg)
	return err
}

// title: list cronjobs
// path: /apps/{appname}/cronjobs
// method: GET
// consume: application/x-www-form-urlencoded
// responses:
//   200: OK
//   400: Invalid data
//   403: Forbidden
//   404: Not found
func listCronjobs(w http.ResponseWriter, r *http.Request, t auth.Token) (err error) {
	appName := r.URL.Query().Get(":appname")
	a, err := getAppFromContext(appName, r)
	if err != nil {
		return err
	}

	canBuild := permission.Check(t, permission.PermAppUpdateCronjobList, contextsForApp(&a)...)
	if !canBuild {
		return &tsuruErrors.HTTP{Code: http.StatusForbidden, Message: "User does not have permission to do this action in this app"}
	}

	evt, err := event.New(&event.Opts{
		Target:     appTarget(appName),
		Kind:       permission.PermAppUpdateCronjobList,
		Owner:      t,
		CustomData: event.FormToCustomData(r.Form),
		Allowed:    event.Allowed(permission.PermAppReadEvents, contextsForApp(&a)...),
	})
	if err != nil {
		return err
	}
	defer func() { evt.Done(err) }()

	msg, err := a.ListCronjobs()
	if err != nil {
		return err
	}

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonMsg)
	return err
}

// title: delete cronjob
// path: /apps/{appname}/cronjobs/{cronjob}
// method: DELETE
// responses:
//   200: Ok
//   401: Unauthorized
//   404: Not found
func deleteCronjob(w http.ResponseWriter, r *http.Request, t auth.Token) (err error) {
	appName := r.URL.Query().Get(":appname")
	a, err := getAppFromContext(appName, r)
	if err != nil {
		return err
	}

	canBuild := permission.Check(t, permission.PermAppUpdateCronjobDelete, contextsForApp(&a)...)
	if !canBuild {
		return &tsuruErrors.HTTP{Code: http.StatusForbidden, Message: "User does not have permission to do this action in this app"}
	}

	evt, err := event.New(&event.Opts{
		Target:     appTarget(appName),
		Kind:       permission.PermAppUpdateCronjobDelete,
		Owner:      t,
		CustomData: event.FormToCustomData(r.Form),
		Allowed:    event.Allowed(permission.PermAppReadEvents, contextsForApp(&a)...),
	})
	if err != nil {
		return err
	}
	defer func() { evt.Done(err) }()

	cronjobName := r.URL.Query().Get(":cronjob")
	err = a.DeleteCronjobs(cronjobName)
	if err != nil {
		return err
	}

	msg := map[string]string{
		"cronjob": cronjobName,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonMsg)
	return err
}

// title: update cronjob
// path: /apps/{appname}/cronjobs/{cronjob}
// method: PUT
// responses:
//   200: Ok
//   401: Unauthorized
//   404: Not found
func updateCronjob(w http.ResponseWriter, r *http.Request, t auth.Token) (err error) {
	err = r.ParseForm()
	if err != nil {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: err.Error()}
	}
	var cronjob provision.CronJob
	dec := form.NewDecoder(nil)
	dec.IgnoreCase(true)
	dec.IgnoreUnknownKeys(true)
	dec.DecodeValues(&cronjob, r.Form)

	appName := r.URL.Query().Get(":appname")
	a, err := getAppFromContext(appName, r)
	if err != nil {
		return err
	}

	canBuild := permission.Check(t, permission.PermAppUpdateCronjobAdd, contextsForApp(&a)...)
	if !canBuild {
		return &tsuruErrors.HTTP{Code: http.StatusForbidden, Message: "User does not have permission to do this action in this app"}
	}

	evt, err := event.New(&event.Opts{
		Target:     appTarget(appName),
		Kind:       permission.PermAppUpdateCronjobAdd,
		Owner:      t,
		CustomData: event.FormToCustomData(r.Form),
		Allowed:    event.Allowed(permission.PermAppReadEvents, contextsForApp(&a)...),
	})
	if err != nil {
		return err
	}
	defer func() { evt.Done(err) }()

	name, err := a.UpdateCronjob(cronjob)
	if err != nil {
		return err
	}

	msg := map[string]string{
		"cronjob": name,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonMsg)
	return err
}
