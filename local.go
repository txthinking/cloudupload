package cloudupload

import (
	"io"
	"os"
	"strings"
)

type Local struct {
	StoragePath string
}

func (l *Local) Save(name string, r io.Reader) error {
	i := strings.LastIndex(name, "/")
	if i != -1 {
		if err := os.MkdirAll(l.StoragePath+"/"+name[0:i], 0740); err != nil {
			return err
		}
	}
	var f *os.File
	f, err := os.OpenFile(l.StoragePath+"/"+name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
