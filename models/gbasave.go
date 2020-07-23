package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/dgrijalva/jwt-go"
)

const (
	// GBADB - db name used for mongo
	GBADB = "GBA SAVE MONGODB NAME"
)

type JWTCustomClaims struct {
	Name  string        `json:"name"`
	Admin bool          `json:"admin"`
	ID    bson.ObjectId `json:"id"`
	Exp   int64         `json:"exp"`
	jwt.StandardClaims
}

type GBASave struct {
	ID         string `json:"id" bson:"_id"`
	Owner      string `json:"owner" bson:"owner"`
	Name       string `json:"name" bson:"name"`
	Save       []byte `json:"save" bson:"save"`
	LastUpdate int64  `json:"last_update" bson:"last_update,omitempty"`
}

func NewGBASave(owner, name string, save []byte) *GBASave {
	id := owner + "_" + name
	return &GBASave{
		ID:         id,
		Owner:      owner,
		Name:       name,
		Save:       save,
		LastUpdate: time.Now().Unix(),
	}
}
