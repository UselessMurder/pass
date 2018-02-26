package routes

import (
	"../dbwrapper"
	"../models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

func getConfigForm(db *dbwrapper.DataBaseWrapper) *form {
	rF := form{}
	rF.Inputs = make([]input, 3)
	rF.Inputs[0].Id, rF.Inputs[0].Name = "min-password-age", "min-password-age"
	rF.Inputs[0].Label = "Минимальный срок действия пароля(секунды):"
	rF.Inputs[0].PlaceHolder = "180"
	rF.Inputs[0].Type = "number"
	rF.Inputs[1].Id, rF.Inputs[1].Name = "max-password-age", "max-password-age"
	rF.Inputs[1].Label = "Максимальный срок действия пароля(секунды):"
	rF.Inputs[1].PlaceHolder = "600"
	rF.Inputs[1].Type = "number"
	rF.Inputs[2].Id, rF.Inputs[2].Name = "max-length-used-list", "max-length-used-list"
	rF.Inputs[2].Label = "Длина списка использованных паролей:"
	rF.Inputs[2].PlaceHolder = "5"
	rF.Inputs[2].Type = "number"
	rF.Send.Value = "Применить"
	row, err := db.QueryRow("GetConfig")
	if err != nil {
		log.Fatal(err)
	}
	var config models.Config
	err = row.Scan(&config.Id, &config.MinPasswordAge, &config.MaxPasswordAge, &config.MaxLengthUsedList)
	if err != nil {
		log.Fatal(err)
	}
	rF.Inputs[0].Value = strconv.Itoa(config.MinPasswordAge)
	rF.Inputs[1].Value = strconv.Itoa(config.MaxPasswordAge)
	rF.Inputs[2].Value = strconv.Itoa(config.MaxLengthUsedList)
	return &rF
}

func GetConfigHandler(c *gin.Context) {
	intr, _ := c.Get("database")
	db := intr.(*dbwrapper.DataBaseWrapper)
	pc := &pageContent{&msg{}, getNavBarWithState(3), getConfigForm(db)}
	c.HTML(http.StatusOK, "common", pc)
}

func PostConfigHandler(c *gin.Context) {
	intr, _ := c.Get("database")
	db := intr.(*dbwrapper.DataBaseWrapper)
	pc := &pageContent{&msg{}, getNavBarWithState(3), getConfigForm(db)}
	minPasswordAgeStr := c.PostForm("min-password-age")
	maxPasswordAgeStr := c.PostForm("max-password-age")
	maxLengthUsedListStr := c.PostForm("max-length-used-list")
	minPasswordAge, maxPasswordAge, maxLengthUsedList := 0, 0, 0
	var err error
	ok, _ := regexp.MatchString(`[1-9]{1,9}`, minPasswordAgeStr)
	minPasswordAge, err = strconv.Atoi(minPasswordAgeStr)
	if !ok || err != nil {
		pc.Message.setError("Некорректный минимальный срок действия пароля!")
		pc.Form.setValues(minPasswordAgeStr, maxPasswordAgeStr, maxLengthUsedListStr)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	ok, _ = regexp.MatchString(`[1-9]{1,9}`, maxPasswordAgeStr)
	maxPasswordAge, err = strconv.Atoi(maxPasswordAgeStr)
	if !ok || err != nil {
		pc.Message.setError("Некорректный максимальный срок действия пароля!")
		pc.Form.setValues(minPasswordAgeStr, maxPasswordAgeStr, maxLengthUsedListStr)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	ok, _ = regexp.MatchString(`[1-9]{1,9}`, maxLengthUsedListStr)
	maxLengthUsedList, err = strconv.Atoi(maxLengthUsedListStr)
	if !ok || err != nil {
		pc.Message.setError("Некорректная длина списка использованных паролей!")
		pc.Form.setValues(minPasswordAgeStr, maxPasswordAgeStr, maxLengthUsedListStr)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	if maxPasswordAge <= minPasswordAge {
		pc.Message.setError("Минимальный срок действия пароля не может быть больше или равен максимальному!")
		pc.Form.setValues(minPasswordAgeStr, maxPasswordAgeStr, maxLengthUsedListStr)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	err = db.ExecTransact("ChangeConfig", minPasswordAge, maxPasswordAge, maxLengthUsedList)
	if err != nil {
		log.Fatal(err)
	}
	pc.Message.setSuccess("Настройки успешно изменены!")
	pc.Form.setValues(minPasswordAgeStr, maxPasswordAgeStr, maxLengthUsedListStr)
	c.HTML(http.StatusOK, "common", pc)
}
