package cloudupload

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Aliyun struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
}

func (a *Aliyun) Save(name string, r io.Reader) error {
	client, err := oss.New(a.Endpoint, a.AccessKeyID, a.AccessKeySecret)
	if err != nil {
		return err
	}
	bucket, err := client.Bucket(a.Bucket)
	if err != nil {
		return err
	}
	if err := bucket.PutObject(name, r); err != nil {
		return err
	}
	return nil
}
