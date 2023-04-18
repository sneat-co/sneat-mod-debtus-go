package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func NewAwsSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"), // get from your AWS console, click "Properties"
		Credentials: credentials.NewStaticCredentials("AKIAIT2ZJZOT2CKJ2JFQ", "BLKRPD57cTtPfczDE2dEu7IgJu/6OpzbA8N+1khN", ""),
	})
}
