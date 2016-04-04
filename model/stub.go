package model

import ()

type Stub struct {
	ID         string `json:"_id"`
	Type       string `json:"type"`
	ParentType string `json:"$ParentType"`
}
