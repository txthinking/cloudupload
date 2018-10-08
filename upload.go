package cloudupload

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/juju/ratelimit"
	uuid "github.com/satori/go.uuid"
	"github.com/txthinking/ant"
)

type Upload struct {
	URL    string
	Stores []Storer
	Rate   int64
}

func (u *Upload) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	uid, err := uuid.NewV4()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	id := strings.Replace(uid.String(), "-", "", -1)
	i := binary.BigEndian.Uint64([]byte(id))
	id = strconv.FormatUint(i, 36)

	var name string
	var b []byte
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		f, fh, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer f.Close()
		b, err = ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		name = fh.Filename
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		var src io.Reader = r.Body
		if u.Rate != 0 {
			bucket := ratelimit.NewBucketWithRate(float64(u.Rate), u.Rate)
			src = ratelimit.Reader(r.Body, bucket)
		}
		b, err = ioutil.ReadAll(src)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if r.Header.Get("Content-Type") == "application/base64" {
			b, err = base64.StdEncoding.DecodeString(string(b))
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
		}
	}
	if name == "" {
		name = Name(r)
	}

	e := make(chan error)
	for _, store := range u.Stores {
		go func(store Storer) {
			e <- store.Save(id+"/"+name, bytes.NewReader(b))
		}(store)
	}
	done := make(chan error)
	var times int
	go func() {
		var isDone bool
		for {
			err := <-e
			if err != nil {
				if !isDone {
					done <- err
					isDone = true
				}
			}
			times++
			if times == len(u.Stores) {
				close(e)
				break
			}
		}
		if !isDone {
			done <- nil
		}
	}()

	err = <-done
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if r.Header.Get("Accept") == "application/json" {
		ant.JSON(w, map[string]string{
			"file": u.URL + id + "/" + ant.URIEscape(name),
		})
		return
	}
	w.Write([]byte(u.URL + id + "/" + ant.URIEscape(name)))
}

func Name(r *http.Request) string {
	name := r.Header.Get("X-File-Name")
	if name != "" {
		s, err := ant.URIUnescape(name)
		if err != nil {
			name = ""
		} else {
			name = s
		}
	}
	if name == "" {
		name = "NoName"
	}
	return name
}
