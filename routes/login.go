package routes

import (
	"../dbwrapper"
	"../hashgenerator"
	"../models"
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func getLoginForm() *form {
	rF := form{}
	rF.Inputs = make([]input, 2)
	rF.Inputs[0].Id, rF.Inputs[0].Name = "login", "login"
	rF.Inputs[0].Label = "Логин:"
	rF.Inputs[0].PlaceHolder = "login"
	rF.Inputs[0].Type = "text"
	rF.Inputs[1].Id, rF.Inputs[1].Name = "password", "password"
	rF.Inputs[1].Label = "Пароль:"
	rF.Inputs[1].PlaceHolder = "password"
	rF.Inputs[1].Type = "password"
	rF.Send.Value = "Вход"
	return &rF
}

func GetLoginHandler(c *gin.Context) {
	pc := &pageContent{&msg{}, getNavBarWithState(1), getLoginForm()}
	c.HTML(http.StatusOK, "common", pc)
}

func PostLoginHandler(c *gin.Context) {
	pc := &pageContent{&msg{}, getNavBarWithState(1), getLoginForm()}
	login := c.PostForm("login")
	password := c.PostForm("password")
	intr, _ := c.Get("database")
	db := intr.(*dbwrapper.DataBaseWrapper)
	row, err := db.QueryRow("GetAccountByName", login)
	if err != nil {
		log.Fatal(err)
	}
	var Acc models.Account
	err = row.Scan(&Acc.Id, &Acc.Name, &Acc.Password, &Acc.PasswordUpdateTime)
	if err != nil {
		pc.Message.setError("Неверный логин или пароль!")
		pc.Form.setValues(login, password)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	hash, _ := hashgenerator.GetHashSum28(password, "magic")
	if bytes.Compare(hash, Acc.Password) != 0 {
		pc.Message.setError("Неверный логин или пароль!")
		pc.Form.setValues(login, password)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	row, err = db.QueryRow("GetConfig")
	if err != nil {
		log.Fatal(err)
	}
	var config models.Config
	err = row.Scan(&config.Id, &config.MinPasswordAge, &config.MaxPasswordAge, &config.MaxLengthUsedList)
	if err != nil {
		log.Fatal(err)
	}
	diff := time.Now().Unix() - Acc.PasswordUpdateTime.Unix()
	if diff > int64(config.MinPasswordAge) {
		if diff > int64(config.MaxPasswordAge) {
			pc.Message.setError("Вход невозможен, пароль сильно устарел!")
			pc.Form.setValues(login, password)
			c.HTML(http.StatusOK, "common", pc)
			return
		}
		pc.Message.setWarning("Успешный вход, но пароль устарел!")
		pc.Form.setValues(login, password)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	pc.Message.setSuccess("Успешный вход!")
	pc.Form.setValues(login, password)
	c.HTML(http.StatusOK, "common", pc)
}
