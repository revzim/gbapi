package models

import (
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Email    string        `json:"email" bson:"email"`
	Password string        `json:"password,omitempty" bson:"password"`
	UserName string        `json:"username,omitempty" bson:"username"`
	Token    string        `json:"token,omitempty" bson:"-"`
	Saves    []GBASave     `json:"saves,omitempty" bson:"saves,omitempty"`
}

func NewUser(email, pw string) *User {
	return &User{
		ID:       bson.NewObjectId(),
		Email:    email,
		Password: pw,
		Token:    "",
		Saves:    []GBASave{},
	}
}
