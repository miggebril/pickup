package models

import (
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"encoding/base64"
	"pickup/helpers"
)

type PlayerStats struct {
	Points int32
	Rebounds int32
	Assists int32
	Blocks int32
	Steals int32
	Turnovers int32
	TotalShotsTaken int32
	TotalShotsMade int32
	ThreePointersTaken int32
	ThreePointersMade int32
}

type BoxScore map[string]PlayerStats

type Result struct {
	HomeScore int32
	AwayScore int32
}

type Game struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty" col:"games"`
	Court 	 bson.ObjectId `json:"-"`
	Name 	 string
	HomeTeam []User
	AwayTeam []User
	Result
	BoxScore map[string]interface{}
}

func (g Game) GetIDEncoded() string {
	return base64.URLEncoding.EncodeToString([]byte(g.ID))
}

func (g *Game) MarshalJSON() ([]byte, error) {
	type Alias Game
	return json.Marshal(&struct {
		ID 				string `json:"ID"`
		Timestamp		string `json:"Timestamp"`
		*Alias
	}{
		ID: helpers.GetIDEncoded(g.ID),
		Timestamp: g.ID.Time().Format("01/02/2006"),
		Alias: (*Alias)(g),
	})
}