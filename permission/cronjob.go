package permission

var (
	PermAppUpdateCronjobAdd    = PermissionRegistry.get("app.update.cronjob.add")    // [global app team pool]
	PermAppUpdateCronjobDelete = PermissionRegistry.get("app.update.cronjob.delete") // [global app team pool]
	PermAppUpdateCronjobList   = PermissionRegistry.get("app.update.cronjob.list")   // [global app team pool]
)
