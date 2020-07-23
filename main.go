package main

import (
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/revzim/gbapi/handlers"
	"github.com/revzim/gbapi/models"
	"gopkg.in/mgo.v2"
)

/*

	db.createUser({
	user: "<NAME>",
	pwd: "<PWD>",
	roles: [{role: "<ROLE_PERMISSION>", db: "<DB_NAME>"}]})
	EX:
		db.createUser({
		user: "TEST_USER",
		pwd: "PASSWORD",
		roles: [{role: "readWrite", db: "AZGBA_DB"}]})
*/
const (
	hosts           = "0.0.0.0"
	dialdb          = "admin"
	username        = ""
	pw              = ""
	gbasaveUsername = ""
	gbasavePWD      = ""
	JWTKeyStr       = ""
)

var ()

func spinupDB() (*mgo.Session, error) {

	// INIT DB CONN
	dbDialInfo := &mgo.DialInfo{
		Addrs:    []string{hosts},
		Timeout:  60 * time.Second,
		Database: dialdb,
		Username: username,
		Password: pw,
	}

	// CONNECT TO DB
	session, err := mgo.DialWithInfo(dbDialInfo)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	if err = session.DB(models.GBADB).Login(gbasaveUsername, gbasavePWD); err != nil {
		panic(err)
	}

	// CREATE INDICES
	if err = session.Copy().DB(models.GBADB).C("users").
		EnsureIndex(mgo.Index{
			Key:    []string{"email"},
			Unique: true,
		}); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return session, nil

}

func main() {

	e := echo.New()

	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.Logger())

	db, err := spinupDB()
	if err != nil {
		log.Fatal(err)
	}

	jwtKey := []byte(JWTKeyStr)

	h := handlers.New(jwtKey, db)

	skipperFunc := func(c echo.Context) bool {
		// SKIP AUTH FOR LOGIN & SIGNUP
		if c.Path() == "/login" || c.Path() == "/signup" {
			return true
		}
		return false
	}

	// CORS CFG
	corsCfg := middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins: []string{"http://azimu", "http://azimu:8080", "http://azimu:8081"},
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderAccessControlAllowOrigin},
	})

	e.Use(corsCfg)

	// JWT MIDDLEWARE ROUTING CFG

	jwtCfg := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(handlers.Key),
		Skipper:    skipperFunc,
	})

	e.Use(jwtCfg)

	// ROUTES
	e.POST("/signup", h.Signup)
	e.POST("/login", h.Login)

	// GBA SAVES ROUTES
	e.POST("/saves/new/:name", h.CreateSave)
	e.POST("/saves/upd/:name", h.UpdateSave)

	e.POST("/saves/upd2/:name", h.UpsertSave)

	e.GET("/saves/game/:name", h.FetchSave)
	e.GET("/saves", h.FetchAllSaves)

	// START
	e.Logger.Fatal(e.Start(":8081"))
}
