package aws_service

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/cjlapao/locally-cli/entities"
)

var globalInstance *AwsService
var mu sync.Mutex

type AwsService struct {
}

func New() *AwsService {
	result := AwsService{}

	return &result
}

func Get() *AwsService {
	mu.Lock()
	defer mu.Unlock()

	if globalInstance == nil {
		globalInstance = New()
	}
	return globalInstance
}

func (c AwsService) Name() string {
	return "aws"
}

func (c *AwsService) NewSession(awsCredentials entities.AwsCredentials) (*session.Session, error) {
	session, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(awsCredentials.Region),
			Credentials: credentials.NewStaticCredentials(awsCredentials.KeyId, awsCredentials.KeySecret, ""),
		},
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (c *AwsService) TestConnection(credentials entities.AwsCredentials) error {
	session, err := c.NewSession(credentials)
	if err != nil {
		return err
	}

	svc := sts.New(session)

	_, err = svc.GetCallerIdentity(nil)
	if err != nil {
		return err
	}

	return nil
}
