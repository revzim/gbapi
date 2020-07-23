package handlers

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/revzim/gbapi/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) Signup(c echo.Context) (err error) {

	// INIT USER
	ctxUser := &models.User{ID: bson.NewObjectId()}

	// BIND USER
	if err = c.Bind(ctxUser); err != nil {
		return
	}

	// VALIDATE USER
	if ctxUser.Email == "" || ctxUser.Password == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid email or password"}
	}

	// SAVE USER TO DB
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(models.GBADB).C("users").Insert(ctxUser); err != nil {
		return c.JSON(http.StatusConflict, echo.Map{
			"error": NewErr(err.Error()),
		})
	}

	return c.JSON(http.StatusCreated, ctxUser)

}

func (h *Handler) Login(c echo.Context) (err error) {

	// c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "http://azimu:8080")
	// INIT USER
	ctxUser := new(models.User)
	if err = c.Bind(ctxUser); err != nil {
		return
	}

	// DB FIND USER
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(models.GBADB).C("users").
		Find(bson.M{"email": ctxUser.Email, "password": ctxUser.Password}).One(ctxUser); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid email or password"}
		}
		return
	}

	// SET CLAIMS
	expTime := time.Duration(15)
	claims := &models.JWTCustomClaims{
		"",
		false,
		ctxUser.ID,
		time.Now().Add(time.Second * expTime).Unix(),
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * expTime).Unix(),
		},
	}

	// HANDLE JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// claims["id"] = ctxUser.ID
	// claims["exp"] = time.Now().Add(time.Second * 600).Unix()

	// INIT ENCODED TOKEN & SEND RESP
	ctxUser.Token, err = token.SignedString([]byte(Key))
	if err != nil {
		return err
	}

	// STRIP PW FOR SAFETY
	ctxUser.Password = ""

	return c.JSON(http.StatusOK, ctxUser)

}

func (h *Handler) Follow(c echo.Context) (err error) {
	// GET USER ID FROM CTX TOKEN
	userID := userIDFromToken(c)

	// GET ID FROM PARAM
	id := c.Param("id")

	// ADD NEW ID AS FOLLOWER TO USER
	db := h.DB.Clone()
	defer db.Close()

	if err = db.DB(models.GBADB).C("users").
		UpdateId(bson.ObjectIdHex(id), bson.M{"$addToSet": bson.M{"followers": userID}}); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
	}

	return

}

func userIDFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}
