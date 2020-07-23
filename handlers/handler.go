package handlers

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	Key []byte
)

type (
	Handler struct {
		DB *mgo.Session
	}
	HandlerErr struct {
		Code   string `json:"code"`
		Reason string `json:"reason"`
	}
)

func New(key []byte, db *mgo.Session) *Handler {
	Key = key
	return &Handler{
		DB: db,
	}
}

func NewErr(msg string) *HandlerErr {
	msgArr := formatErrorString(msg)
	return &HandlerErr{
		Code:   msgArr[0],
		Reason: msgArr[1],
	}
}

func formatErrorString(errStr string) []string {
	errSplit := strings.Split(errStr, " collection:")[0]
	errs := strings.SplitN(errSplit, " ", 2)
	code := errs[0]
	reason := fmt.Sprintf("%s", errs[1])
	return []string{code, reason}
}

func (h *Handler) HandleUpsert(saveData, owner string) map[string]bson.M {
	// UPDATED QUERY VAL
	update := bson.M{"$set": bson.M{"save": saveData}}
	// WHAT TO MATCH QUERY WITH FOR UPSERT
	selector := bson.M{"owner": owner}
	return map[string]bson.M{
		"selector": selector,
		"update":   update,
	}
}

func (h *Handler) InitUpdateSave(saveData []byte) bson.M {
	return bson.M{"$set": bson.M{"save": saveData, "last_update": time.Now().Unix()}}
}

func (h *Handler) InitUpsertSave(id, name, owner string, saveData []byte) bson.M {
	return bson.M{"$set": bson.M{"id": owner + "_" + name, "owner": owner, "last_update": time.Now().Unix(), "name": name, "save": saveData}}
}

func (h *Handler) InitUpsertIDMsg(saveData []byte) bson.M {
	return bson.M{"$set": bson.M{"save": saveData, "last_update": time.Now().Unix()}}
}
