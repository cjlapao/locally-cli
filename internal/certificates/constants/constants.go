// Package constants provides the constants for the certificates.
package constants

import "encoding/asn1"

var (
	OidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}

	ServerAuthentication      = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 1}
	ClientAuthentication      = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 2}
	CodeSigning               = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 3}
	SecureEmail               = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 4}
	TimeStamping              = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 8}
	OCSPSigning               = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 9}
	MicrosoftTrustListSigning = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 3, 1}
	EncryptingFileSystem      = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 3, 4}

	Policy4 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3}
	Policy5 = asn1.ObjectIdentifier{2, 23, 140, 1, 2, 1}
	Policy6 = asn1.ObjectIdentifier{2, 23, 140, 1, 2, 2}

	RootPolicy3 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3, 2}
	RootPolicy4 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 5, 3, 3}

	CertificateStorageFolder = "certificates"
)
