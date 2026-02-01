package platform

// StoragePermissionService provides storage/file system permission access.
type StoragePermissionService struct {
	// Permission for storage access.
	Permission Permission
}

// StoragePermission is the singleton storage permission service.
var StoragePermission = &StoragePermissionService{
	Permission: &basicPermission{inner: newPermission("storage")},
}
