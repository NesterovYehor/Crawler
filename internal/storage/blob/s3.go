package blob

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Client struct {
	uploader *s3manager.Uploader
}

func NewS3Client() (BlobStore, error) {
	sess, err := session.NewSession()
	uploader := s3manager.NewUploader(sess)
	return &s3Client{uploader: uploader}, err
}

func (c *s3Client) Save(ctx context.Context, data []byte) error {
	return nil
}
