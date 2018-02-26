package routes

import (
	"../dbwrapper"
	"../hashgenerator"
	"../models"
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"regexp"
	"time"
)

func getChangePasswordForm() *form {
	rF := form{}
	rF.Inputs = make([]input, 4)
	rF.Inputs[0].Id, rF.Inputs[0].Name = "login", "login"
	rF.Inputs[0].Label = "Логин:"
	rF.Inputs[0].PlaceHolder = "login"
	rF.Inputs[0].Type = "text"
	rF.Inputs[1].Id, rF.Inputs[1].Name = "old-password", "old-password"
	rF.Inputs[1].Label = "Старый пароль:"
	rF.Inputs[1].PlaceHolder = "old-password"
	rF.Inputs[1].Type = "password"
	rF.Inputs[2].Id, rF.Inputs[2].Name = "new-password", "new-password"
	rF.Inputs[2].Label = "Новый пароль:"
	rF.Inputs[2].PlaceHolder = "new-password"
	rF.Inputs[2].Type = "password"
	rF.Inputs[3].Id, rF.Inputs[3].Name = "confirm-password", "confirm-password"
	rF.Inputs[3].Label = "Подтвердите пароль:"
	rF.Inputs[3].PlaceHolder = "new-password"
	rF.Inputs[3].Type = "password"
	rF.Send.Value = "Сменить"
	return &rF
}

func GetChangePasswordHandler(c *gin.Context) {
	pc := &pageContent{&msg{}, getNavBarWithState(2), getChangePasswordForm()}
	c.HTML(http.StatusOK, "common", pc)
}

func PostChangePasswordHandler(c *gin.Context) {
	pc := &pageContent{&msg{}, getNavBarWithState(2), getChangePasswordForm()}
	login := c.PostForm("login")
	oldPassword := c.PostForm("old-password")
	newPassword := c.PostForm("new-password")
	confirmPassword := c.PostForm("confirm-password")
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
		pc.Form.setValues(login, oldPassword, newPassword, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	hash, _ := hashgenerator.GetHashSum28(oldPassword, "magic")
	if bytes.Compare(hash, Acc.Password) != 0 {
		pc.Message.setError("Неверный логин или пароль!")
		pc.Form.setValues(login, oldPassword, newPassword, confirmPassword)
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
	if diff <= int64(config.MinPasswordAge) {
		pc.Message.setWarning("Менять пароль еще рано!")
		pc.Form.setValues(login, oldPassword, newPassword, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	ok, _ := regexp.MatchString(`[а-яёА-ЯЁ[:graph:]]{3,30}`, newPassword)
	if !ok {
		pc.Message.setError("Новый пароль некорректен!")
		pc.Form.setValues(login, oldPassword, newPassword, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	hash, _ = hashgenerator.GetHashSum28(newPassword, "magic")
	if !validatePassword(hash, db) {
		pc.Message.setError("Используйте другой новый пароль!")
		pc.Form.setValues(login, oldPassword, newPassword, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	if newPassword != confirmPassword {
		pc.Message.setError("Пароли не совпадают!")
		pc.Form.setValues(login, oldPassword, newPassword, confirmPassword)
		c.HTML(http.StatusOK, "common", pc)
		return
	}
	addPassword(hash, db)
	err = db.ExecTransact("ChangePassword", hash, time.Now(), Acc.Id)
	if err != nil {
		log.Fatal(err)
	}
	pc.Message.setSuccess("Пароль успешно изменен!")
	pc.Form.setValues(login, oldPassword, newPassword, confirmPassword)
	c.HTML(http.StatusOK, "common", pc)
}
