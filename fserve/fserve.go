package fserve

import (
	// 	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/flimzy/go-pouchdb"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"

	"github.com/flimzy/flashback-model"
	"github.com/flimzy/flashback/repository"
)

func Init(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		js.Global.Call("addEventListener", "message", func(e *js.Object) {
			data := e.Get("data").String()
			var msg Message
			if err := json.Unmarshal([]byte(data), &msg); err != nil {
				fmt.Printf("Error decoding message from iframe: %s\n", err)
				return
			}
			go func() {
				data, err := fetchFile(&msg)
				if err != nil {
					fmt.Printf("Error fetching file: %s\n", err)
					return
				}
				if err := sendResponse(&msg, data); err != nil {
					fmt.Printf("Error sending response to iframe: %s\n", err)
					return
				}
			}()
		})
	}()
}

func fetchFile(req *Message) (*string, error) {
	for _, id := range []string{req.NoteId, req.ModelId} {
		file, err := fetchAttachment(id, req.Path)
		if file != nil || err != nil {
			return file, err
		}
	}
	switch req.Path {
	case "script.js":
		return encodeFile("text/javascript", []byte("console.log('placeholder'); /* JS placeholder */")), nil
	case "style.css":
		return encodeFile("text/css", []byte("/* CSS placeholder */")), nil
	}
	return nil, fmt.Errorf("File not found: %s", req.Path)
}

func sendResponse(req *Message, data *string) error {
	iframe := js.Global.Get("document").Call("getElementById", req.IframeId)
	if jsbuiltin.TypeOf(iframe) == "undefined" {
		return fmt.Errorf("Cannot find requested iframe")
	}
	iframe.Get("contentWindow").Call("postMessage", Response{
		Tag:  req.Tag,
		Path: req.Path,
		Data: *data,
	}, "*")
	return nil
}

func encodeFile(contentType string, data []byte) *string {
	b64data := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)
	return &b64data
}

func fetchAttachment(id, filename string) (*string, error) {
	u, err := repo.CurrentUser()
	if err != nil {
		return nil, err
	}
	db, err := u.DB()
	if err != nil {
		return nil, err
	}

	var note fb.Note
	if err := db.Get(id, &note, pouchdb.Options{}); err != nil {
		return nil, fmt.Errorf("Error fetching note: %s\n", err)
	}
	// 	att, ok := note.Attachments.GetFile(filename)
	// 	for attName, attInfo := range note.Attachments {
	// 		if attName == filename {
	// 			att, err := db.Attachment(note.Id, filename, note.Rev)
	// 			if err != nil {
	// 				return nil, fmt.Errorf("Error fetching attachment '%s' from note: %s\n", attName, err)
	// 			}
	// 			buf := new(bytes.Buffer)
	// 			buf.ReadFrom(att.Body)
	// 			return encodeFile(attInfo.Type, buf.Bytes()), nil
	// 		}
	// 	}
	return nil, nil
}

type Message struct {
	IframeId string
	Tag      string
	CardId   string
	NoteId   string
	ModelId  string
	Path     string
}

type Response struct {
	Tag  string
	Path string
	Data string
}
