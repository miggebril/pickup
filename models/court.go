package models

import (
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"encoding/base64"
)

type Location struct {
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Court struct {
	ID bson.ObjectId `json:"-" bson:"_id,omitempty" col:"courts"` 
	Name string
	City string
	State string
	Zipcode int32
	Neighborhood string
	Location
	Rating float64
}

func (c Court) GetIDEncoded() string {
	return base64.URLEncoding.EncodeToString([]byte(c.ID))
}

func (c *Court) MarshalJSON() ([]byte, error) {
	type Alias Court
	return json.Marshal(&struct {
		ID 				string `json:"ID"`
		*Alias
	}{
		ID: c.GetIDEncoded(),
		Alias: (*Alias)(c),
	})
}