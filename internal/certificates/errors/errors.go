package errors

// Package errors contains the error codes for the application, this is used to standardize the error codes across the application
const (
	// General errors
	ErrorInternalServer = "CRT001"
	ErrorNotFound       = "CRT002"
	ErrorInvalidRequest = "CRT003"

	// Certificate errors
	ErrorGettingCertificates = "CRT004"
	ErrorCreatingCertificate = "CRT005"
	ErrorUpdatingCertificate = "CRT006"
	ErrorDeletingCertificate = "CRT007"

	// Pagination errors
	ErrorParsingPaginationRequest = "CRT011"

	// Tenant errors
	ErrorMissingTenantID = "CRT012"

	// Root certificate errors
	ErrorGettingRootCertificate  = "CRT013"
	ErrorCreatingRootCertificate = "CRT014"
	ErrorUpdatingRootCertificate = "CRT015"
	ErrorDeletingRootCertificate = "CRT016"
)
