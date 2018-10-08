package cloudupload

import "io"

type Storer interface {
	Save(name string, r io.Reader) error
}
