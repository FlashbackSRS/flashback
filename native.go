// +build !js

package flashback

import (
	"bytes"
	"log"
	"net/http"
)

func Console(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func (f *FlashbackClient) Get(uri string) (string, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()
	return body, nil
}
