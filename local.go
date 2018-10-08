package cloudupload

import (
	"io"
	"os"
)

type Local struct {
	StoragePath string
}

func (l *Local) Save(name string, r io.Reader) error {
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
