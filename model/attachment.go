package model

type Attachment struct {
	ContentType string `json:"content-type"`
	Data        []byte `json:"data"`
}
