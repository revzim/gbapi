package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/revzim/gbapi/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) CreateSave(c echo.Context) (err error) {
	// INIT USER
	ctxUser := &models.User{
		ID: bson.ObjectIdHex(userIDFromToken(c)),
	}

	// CREATE NEW GRIDFS FOR SAVE FILE
	// HANDLE SAVE FILE
	file, err := c.FormFile("save_file")
	if err != nil {
		return err
	}

	fileBytes, err := readSave(file)
	if err != nil {
		return err
	}

	// fmt.Printf("save file size==> %d bytes", len(fileBytes))
	// c.Logger().Debugf("save file size==> %d bytes", len(fileBytes))

	// INIT SAVE
	save := &models.GBASave{
		ID:         ctxUser.ID.Hex() + "_" + c.Param("name"),
		Owner:      ctxUser.ID.Hex(),
		LastUpdate: time.Now().Unix(),
		Name:       c.Param("name"),
		Save:       fileBytes,
	}
	// ATTEMPT BIND SAVE
	// if err = c.Bind(save); err != nil {
	// 	return
	// }

	// fmt.Println("save:", save)

	// VALIDATE
	if save.Name == "" || len(save.Save) < 1 {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid recipient or message fields"}
	}

	// ATTEMPT FIND USER IN DB
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(models.GBADB).C("users").FindId(ctxUser.ID).One(ctxUser); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return err
	}

	// INSERT USER SAVE IN DB
	if err = db.DB(models.GBADB).C("saves").Insert(save); err != nil {
		return
	}

	c.Logger().Debugf("success save file : size==> %d bytes", len(fileBytes))

	// SAVE INSERT INTO DB
	return c.JSON(http.StatusCreated, save)

}

func readSave(f *multipart.FileHeader) ([]byte, error) {

	saveFileSrc, err := f.Open()
	if err != nil {
		return nil, err
	}

	defer saveFileSrc.Close()

	data := make([]byte, f.Size)

	_, err = saveFileSrc.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil

}

func initGridFile(fileBytes []byte, db *mgo.Database, filename string) (*mgo.GridFile, error) {

	saveFile, err := db.GridFS("fs").Create(filename)
	if err != nil {
		return nil, err
	}

	// WRITE SAVE FILE DATA TO FILE
	saveFileSize, err := saveFile.Write(fileBytes)
	if err != nil {
		return nil, err
	}
	// CLOSE FILE
	err = saveFile.Close()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Created GridFS file:%s size: %d | write size: %d\n", saveFile.Name(), saveFile.Size(), saveFileSize)
	return saveFile, nil
}

func (h *Handler) UpdateSave(c echo.Context) (err error) {
	// INIT USER
	ctxUser := &models.User{
		ID: bson.ObjectIdHex(userIDFromToken(c)),
	}

	// CREATE NEW GRIDFS FOR SAVE FILE
	// HANDLE SAVE FILE
	file, err := c.FormFile("save_file")
	if err != nil {
		return err
	}

	fileBytes, err := readSave(file)
	if err != nil {
		return err
	}
	//c.Logger().Debugf("save file size==> %d bytes", len(fileBytes))
	// fmt.Printf("save file size==> %d bytes", len(fileBytes))

	// VALIDATE
	if c.Param("name") == "" || len(fileBytes) < 1 {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid save file data or file name"}
	}

	// ATTEMPT FIND USER IN DB
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(models.GBADB).C("users").FindId(ctxUser.ID).One(ctxUser); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return err
	}
	// UPSERTID SAVE IN DB
	saveID := ctxUser.ID.Hex() + "_" + c.Param("name")
	update := h.InitUpdateSave(fileBytes)
	//
	if err = db.DB(models.GBADB).C("saves").
		UpdateId(saveID, update); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return
	}

	//fmt.Println("\nUPDATED SAVE WITH NEW DATA!")
	c.Logger().Debugf("updated save : old | new file size ==> %d bytes | %d bytes", len(fileBytes))
	return c.JSON(http.StatusOK, echo.Map{
		"user_id": ctxUser.ID.Hex(),
		"game":    c.Param("name"),
		"updated": true,
	})

}

func (h *Handler) UpsertSave(c echo.Context) (err error) {
	// INIT USER
	ctxUser := &models.User{
		ID: bson.ObjectIdHex(userIDFromToken(c)),
	}

	// CREATE NEW GRIDFS FOR SAVE FILE
	// HANDLE SAVE FILE
	file, err := c.FormFile("save_file")
	if err != nil {
		return err
	}

	fileBytes, err := readSave(file)
	if err != nil {
		return err
	}
	//c.Logger().Debugf("save file size==> %d bytes", len(fileBytes))
	// fmt.Printf("save file size==> %d bytes", len(fileBytes))

	// VALIDATE
	if c.Param("name") == "" || len(fileBytes) < 1 {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid save file data or file name"}
	}

	// ATTEMPT FIND USER IN DB
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB(models.GBADB).C("users").FindId(ctxUser.ID).One(ctxUser); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return err
	}
	// UPSERTID SAVE IN DB
	saveID := ctxUser.ID.Hex() + "_" + c.Param("name")
	update := h.InitUpsertSave(saveID, c.Param("name"), ctxUser.ID.Hex(), fileBytes)
	//
	info, err := db.DB(models.GBADB).C("saves").
		UpsertId(saveID, update)

	if err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return err
	}

	fmt.Println(info)

	//fmt.Println("\nUPDATED SAVE WITH NEW DATA!")
	c.Logger().Debugf("updated save : new file size ==> %d bytes", len(fileBytes))
	return c.JSON(http.StatusOK, echo.Map{
		"user_id": ctxUser.ID.Hex(),
		"game":    c.Param("name"),
		"updated": true,
	})
}

func (h *Handler) FetchAllSaves(c echo.Context) (err error) {
	// c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "http://azimu:8080")
	// println("access:", c.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))
	// GET ID FROM TOKEN
	userID := userIDFromToken(c)

	// ATTEMPT TO GET GBASaves FROM DB
	saves := []*models.GBASave{}
	db := h.DB.Clone()
	if err = db.DB(models.GBADB).C("saves").
		Find(bson.M{"owner": userID}).
		All(&saves); err != nil {
		return
	}

	defer db.Close()

	return c.JSON(http.StatusOK, saves)

}

func (h *Handler) FetchSave(c echo.Context) (err error) {

	// GET ID FROM TOKEN
	userID := userIDFromToken(c)

	// GET SAVE NAME FROM FORM
	name := c.Param("name")

	// ATTEMPT TO GET GBASaves FROM DB
	saves := []*models.GBASave{}
	db := h.DB.Clone()
	if err = db.DB(models.GBADB).C("saves").
		Find(bson.M{"owner": userID, "name": name}).
		All(&saves); err != nil {
		return
	}

	defer db.Close()

	return c.JSON(http.StatusOK, saves)

}
