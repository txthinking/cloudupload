package cloudupload

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/txthinking/x"
)

type Tencent struct {
	SecretId  string
	SecretKey string
	Host      string
}

func (t *Tencent) Save(name string, r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	ss := strings.Split(name, "/")
	rq, err := http.NewRequest("PUT", "https://"+t.Host+"/"+ss[0]+"/"+x.URIEscape(ss[1]), bytes.NewReader(data))
	if err != nil {
		return err
	}
	a, err := t.Authorization(name, len(data))
	if err != nil {
		return err
	}
	rq.Header.Set("Authorization", a)
	rq.Header.Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	res, err := c.Do(rq)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}
	return nil
}

func (t *Tencent) Authorization(name string, length int) (string, error) {
	ts0 := time.Now().Unix()
	ts1 := time.Now().Unix() + 60*60
	ts := strconv.FormatInt(ts0, 10) + ";" + strconv.FormatInt(ts1, 10)
	s := ""
	s += "q-sign-algorithm=sha1"
	s += "&q-ak=" + t.SecretId
	s += "&q-sign-time=" + ts
	s += "&q-key-time=" + ts
	s += "&q-header-list=" + "content-length;host"
	s += "&q-url-param-list=" + ""

	mac := hmac.New(sha1.New, []byte(t.SecretKey))
	if _, err := mac.Write([]byte(ts)); err != nil {
		return "", err
	}
	signKey := hex.EncodeToString(mac.Sum(nil))
	httpString := fmt.Sprintf("put\n/%s\n%s\ncontent-length=%d&host=%s\n", name, "", length, t.Host)
	stringToSign := fmt.Sprintf("sha1\n%s\n%s\n", ts, x.SHA1(httpString))
	mac = hmac.New(sha1.New, []byte(signKey))
	if _, err := mac.Write([]byte(stringToSign)); err != nil {
		return "", err
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	s += "&q-signature=" + signature

	return s, nil
}
