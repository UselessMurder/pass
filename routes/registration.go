package routes

import (
	"../dbwrapper"
	"../hashgenerator"
	"../models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"regexp"
	"time"
)

func getRegistrationForm() *form {
	rF := form{}
	rF.Inputs = make([]input, 3)
	rF.Inputs[0].Id, rF.Inputs[0].Name = "login", "login"
	rF.Inputs[0].Label = "Логин:"
	rF.Inputs[0].PlaceHolder = "login"
	rF.Inputs[0].Type = "text"
	rF.Inputs[1].Id, rF.Inputs[1].Name = "password", "password"
	rF.Inputs[1].Label = "Пароль:"
	rF.Inputs[1].PlaceHolder = "password"
	rF.Inputs[1].Type = "password"
	rF.Inputs[2].Id, rF.Inputs[2].Name = "confirm-password", "confirm-password"
	rF.Inputs[2].Label = "Подтвердите пароль:"
	rF.Inputs[2].PlaceHolder = "password"
	rF.Inputs[2].Type = "password"
	rF.Send.Value = "Зарегистрировать"
	return &rF
}

func GetRegistrationHandler(c *gin.Context) {
	pc := &pageContent{&msg{}, getNavBarWithState(0), getRegistrationForm()}
	c.HTML(http.StatusOK, "common", pc)
}

func PostRegistrationHandler(c *gin.Context) {
	pc := &pageContent{&msg{}, getNavBarWithState(0), getRegistrationForm()}
	login := c.PostForm("login")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm-password")
	intr, _ := c.Get("database")
	db := intr.(*dbwrapper.DataBaseWrapper)
	ok, _ := regexp.MatchString(`[a-zA-Zа-яёА-ЯЁ1-90]{3,30}`, login)
	if !ok {
		pc.Message.setError("Некорректный логин!")
		pc.Form.setValues(login, password, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	row, err := db.QueryRow("GetAccountByName", login)
	if err != nil {
		log.Fatal(err)
	}
	var Acc models.Account
	err = row.Scan(&Acc.Id, &Acc.Name, &Acc.Password, &Acc.PasswordUpdateTime)
	if err == nil {
		pc.Message.setError("Логин уже используется!")
		pc.Form.setValues(login, password, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	ok, _ = regexp.MatchString(`[а-яёА-ЯЁ[:graph:]]{3,30}`, password)
	if !ok {
		pc.Message.setError("Некорректный пароль!")
		pc.Form.setValues(login, password, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}

	hash, _ := hashgenerator.GetHashSum28(password, "magic")
	if !validatePassword(hash, db) {
		pc.Message.setError("Используйте другой пароль!")
		pc.Form.setValues(login, password, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	if password != confirmPassword {
		pc.Message.setError("Пароли не совпадают!")
		pc.Form.setValues(login, password, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	addPassword(hash, db)
	err = db.ExecTransact("RegisterUser", login, hash, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	pc.Message.setSuccess("Учетная запись зарегистрирована!")
	pc.Form.setValues(login, password, confirmPassword)
	c.HTML(http.StatusOK, "common", pc)
}
