package types

import (
	"crypto/x509"
	"strconv"
	"strings"
)

type CertificateKeySize int

const (
	CertificateKeySize2048 CertificateKeySize = 2048
	CertificateKeySize4096 CertificateKeySize = 4096
	CertificateKeySize8192 CertificateKeySize = 8192
)

func IntToCertificateKeySize(i int) CertificateKeySize {
	switch i {
	case 2048:
		return CertificateKeySize2048
	case 4096:
		return CertificateKeySize4096
	case 8192:
		return CertificateKeySize8192
	default:
		return CertificateKeySize4096
	}
}

func StringToCertificateKeySize(str string) int {
	if str == "" {
		return 2048
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		return 2048
	}

	return i
}

func (k CertificateKeySize) ToX509KeySize() string {
	return strconv.Itoa(int(k))
}

type SignatureAlgorithm int

const (
	SignatureAlgorithmSHA256 SignatureAlgorithm = 4
	SignatureAlgorithmSHA384 SignatureAlgorithm = 5
	SignatureAlgorithmSHA512 SignatureAlgorithm = 6
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

func IntToSignatureAlgorithm(i int) SignatureAlgorithm {
	switch i {
	case 4:
		return SignatureAlgorithmSHA256
	case 5:
		return SignatureAlgorithmSHA384
	case 6:
		return SignatureAlgorithmSHA512
	default:
		return SignatureAlgorithmSHA256
	}
}

func StringToSignatureAlgorithm(str string) SignatureAlgorithm {
	if str == "" {
		return SignatureAlgorithmSHA512
	}

	str = strings.ToUpper(str)
	switch str {
	case "SHA256":
		return SignatureAlgorithmSHA256
	case "SHA384":
		return SignatureAlgorithmSHA384
	case "SHA512":
		return SignatureAlgorithmSHA512
	}

	return SignatureAlgorithmSHA512
}

func (s SignatureAlgorithm) ToString() string {
	switch s {
	case SignatureAlgorithmSHA256:
		return "SHA256"
	case SignatureAlgorithmSHA384:
		return "SHA384"
	case SignatureAlgorithmSHA512:
		return "SHA512"
	}
	return "SHA256"
}

type CertificateType string

const (
	CertificateTypeRoot         CertificateType = "root"
	CertificateTypeIntermediate CertificateType = "intermediate"
	CertificateTypeCertificate  CertificateType = "certificate"
)
