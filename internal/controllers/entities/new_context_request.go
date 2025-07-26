package entities

import (
	"encoding/json"
	"errors"
)

type LocationType string

const (
	Locally LocationType = "locally"
	Azure   LocationType = "azure"
	Aws     LocationType = "aws"
)

type NewContextRequest struct {
	Name          string                     `json:"name" yaml:"name"`
	LocationType  LocationType               `json:"type" yaml:"type"`
	Locally       *NewContextLocallyLocation `json:"locally,omitempty" yaml:"locally,omitempty"`
	Azure         *NewContextAzureLocation   `json:"azure,omitempty" yaml:"azure,omitempty"`
	Aws           *NewContextAwsLocation     `json:"aws,omitempty" yaml:"aws,omitempty"`
	DomainName    string                     `json:"domainName" yaml:"domainName"`
	SubDomainName string                     `json:"subDomainName" yaml:"subDomainName"`
}

type NewContextResponse struct {
	Id      string `json:"id" yaml:"id"`
	Success bool   `json:"success" yaml:"success"`
}

type NewContextLocallyLocation struct {
	Path string `json:"path" yaml:"path"`
}

type NewContextAzureLocation struct {
	SubscriptionId     string `json:"subscriptionId" yaml:"subscriptionId"`
	TenantId           string `json:"tenantId" yaml:"tenantId"`
	ClientId           string `json:"clientId" yaml:"clientId"`
	ClientSecret       string `json:"clientSecret" yaml:"clientSecret"`
	StorageAccountName string `json:"storageAccountName" yaml:"storageAccountName"`
	ResourceGroupName  string `json:"resourceGroupName" yaml:"resourceGroupName"`
	ContainerName      string `json:"containerName" yaml:"containerName"`
}

type NewContextAwsLocation struct {
	AccessKeyId     string `json:"accessKeyId" yaml:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret" yaml:"accessKeySecret"`
	Region          string `json:"region" yaml:"region"`
	BucketName      string `json:"bucketName" yaml:"bucketName"`
}

func (l *LocationType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	switch s {
	case "locally":
		*l = Locally
	case "azure":
		*l = Azure
	case "aws":
		*l = Aws
	default:
		return errors.New("invalid location type")
	}

	return nil
}

func (l LocationType) MarshalJSON() ([]byte, error) {
	switch l {
	case Locally:
		return json.Marshal("locally")
	case Azure:
		return json.Marshal("azure")
	case Aws:
		return json.Marshal("aws")
	default:
		return nil, errors.New("invalid location type")
	}
}
