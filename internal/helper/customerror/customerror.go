package customerror

import "gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"

// Error groups
const (
	ErrGroupClientErr          = "client-error"
	ErrGroupInternalErr        = "internal-error"
	ErrGroupDataNotFoundErr    = "data-not-found-error"
	ErrGroupForbiddenErr       = "forbidden-error"
	ErrGroupServiceUnavailable = "service-closed"
	ErrGroupUnauthorized       = "unauthorized"
	ErrGroupForbidden          = "forbidden"
)

// Common error codes
const (
	ErrCodeDatabase = "DATABASE_ERR"
)

// Common errors
var (
	ErrDatabase           = errors.New(ErrGroupInternalErr, ErrCodeDatabase, "database operation error")
	ErrInternal           = errors.New(ErrGroupInternalErr, "INTERNAL_ERR", "")
	ErrNotFound           = errors.New(ErrGroupDataNotFoundErr, "NOT_FOUND", "data not found")
	ErrServiceUnavailable = errors.New(ErrGroupServiceUnavailable, "", "Service unavailable")
)

var (
	ErrMemberNotFound = errors.New(ErrGroupDataNotFoundErr, "1", "Member tidak ditemukan")
)
