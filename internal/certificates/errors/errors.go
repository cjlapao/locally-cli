// Package errors provides error codes for the certificates service
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

	// User errors
	ErrorMissingUser = "CRT013"

	// Root certificate errors
	ErrorGettingRootCertificate = "CRT014"

	// Intermediate certificate errors
	ErrorGettingIntermediateCertificate = "CRT015"

	// Certificate type errors
	ErrorMissingCertificateType         = "CRT016"
	ErrorGettingCertificateByType       = "CRT017"
	ErrorGettingCertificatesByType      = "CRT018"
	ErrorMissingCertificateID           = "CRT019"
	ErrorGettingCertificate             = "CRT020"
	ErrorMissingRootCertificate         = "CRT021"
	ErrorMissingIntermediateCertificate = "CRT022"

	// Sub domain errors
	ErrorMissingSubDomain = "CRT023"

	// FQDN errors
	ErrorMissingFQDNs = "CRT024"

	// User ID errors
	ErrorMissingUserID = "CRT025"

	// Access errors
	ErrorDeletingRootCertificate         = "CRT026"
	ErrorDeletingIntermediateCertificate = "CRT027"
)
