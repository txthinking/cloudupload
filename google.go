package cloudupload

import (
	"io"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

type Google struct {
	Bucket string
}

func (g *Google) Save(name string, r io.Reader) error {
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, time.Second*60*5)
	//defer cancel()
	client, err := storage.NewClient(ctx)
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
