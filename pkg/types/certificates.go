package types

import "crypto/x509"

type CertificateKeySize int

const (
	Key2048 CertificateKeySize = 2048
	Key4096 CertificateKeySize = 4096
	Key8192 CertificateKeySize = 8192
)

type SignatureAlgorithm int

const (
	SHA256 SignatureAlgorithm = 4
	SHA384 SignatureAlgorithm = 5
	SHA512 SignatureAlgorithm = 6
)

func (s SignatureAlgorithm) ToX509SignatureAlgorithm() x509.SignatureAlgorithm {
	switch s {
	case 4:
		return x509.SHA256WithRSA
	case 5:
		return x509.SHA384WithRSA
	case 6:
		return x509.SHA512WithRSA
	default:
		return x509.SHA256WithRSA
	}
}
