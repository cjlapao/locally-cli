// Package errors contains the error codes for the application, this is used to standardize the error codes across the application
package errors

const (
	ErrorGettingRootCertificate          = "CRT001"
	ErrorCreatingRootCertificate         = "CRT002"
	ErrorDeletingRootCertificate         = "CRT003"
	ErrorGettingIntermediateCertificate  = "CRT004"
	ErrorCreatingIntermediateCertificate = "CRT005"
	ErrorDeletingIntermediateCertificate = "CRT006"
	ErrorGettingServerCertificate        = "CRT007"
	ErrorCreatingServerCertificate       = "CRT008"
	ErrorGettingClientCertificate        = "CRT009"
	ErrorCreatingClientCertificate       = "CRT010"

	// Pagination errors
	ErrorParsingPaginationRequest = "CRT011"

	// Tenant errors
	ErrorMissingTenantID = "CRT012"
)
