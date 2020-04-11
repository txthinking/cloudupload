package cloudupload

import (
	"io"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

type Google struct {
	ServiceAccountFile string
	Bucket             string
}

func (g *Google) Save(name string, r io.Reader) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(g.ServiceAccountFile))
	if err != nil {
		return err
	}
	bkt := client.Bucket(g.Bucket)
	obj := bkt.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}
