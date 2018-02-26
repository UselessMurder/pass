package routes

import (
	"../dbwrapper"
	"../models"
	"database/sql"
	"log"
)

type pageContent struct {
	Message *msg
	Nav     *navbarContent
	Form    *form
}

type form struct {
	Inputs []input
	Send   submit
}

func (f *form) setValues(values ...string) {
	for key := range f.Inputs {
		f.Inputs[key].Value = values[key]
	}
}

type msg struct {
	Active bool
	Value  string
	Type   uint8
}

func (m *msg) setError(value string) {
	m.Active = true
	m.Value = value
	m.Type = 0
}

func (m *msg) setSuccess(value string) {
	m.Active = true
	m.Value = value
	m.Type = 1
}

func (m *msg) setWarning(value string) {
	m.Active = true
	m.Value = value
	m.Type = 2
}

type link struct {
	Id     string
	Value  string
	Href   string
	Active bool
}

type input struct {
	Id          string
	Name        string
	PlaceHolder string
	Label       string
	Value       string
	Type        string
}

type submit struct {
	Value string
}

type navbarContent struct {
	Links []link
}

func (nC *navbarContent) init() {
	nC.Links = make([]link, 4)
	nC.Links[0].Value = "Регистрация"
	nC.Links[0].Href = "/registration"
	nC.Links[1].Value = "Вход"
	nC.Links[1].Href = "/login"
	nC.Links[2].Value = "Сменить пароль"
	nC.Links[2].Href = "/change_password"
	nC.Links[3].Value = "Настройки"
	nC.Links[3].Href = "/config"
}

func getNavBarWithState(state int) *navbarContent {
	var currentNav navbarContent
	currentNav.init()
	switch state {
	case 0:
		currentNav.Links[0].Active = true
	case 1:
		currentNav.Links[1].Active = true
	case 2:
		currentNav.Links[2].Active = true
	case 3:
		currentNav.Links[3].Active = true
	}
	return &currentNav
}

func validatePassword(hash []byte, db *dbwrapper.DataBaseWrapper) bool {
	row, err := db.QueryRow("GetUsedPasswordByValue", hash)
	if err != nil {
		log.Fatal(err)
	}
	var usedPass models.UsedPassword
	err = row.Scan(&usedPass.Id, &usedPass.Password)
	if err == nil {
		return false
	}
	return true
}

func addPassword(hash []byte, db *dbwrapper.DataBaseWrapper) {
	row, err := db.QueryRow("GetConfig")
	if err != nil {
		log.Fatal(err)
	}
	var config models.Config
	err = row.Scan(&config.Id, &config.MinPasswordAge, &config.MaxPasswordAge, &config.MaxLengthUsedList)
	if err != nil {
		log.Fatal(err)
	}
	err = db.ExecTransact("AddPassword", hash)
	if err != nil {
		log.Fatal(err)
	}
	var rows *sql.Rows
	rows, err = db.Query("GetUsedPasswords")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	lastId, count := 0, 0
	for rows.Next() {
		var usedPass models.UsedPassword
		rows.Scan(&usedPass.Id, &usedPass.Password)
		lastId = usedPass.Id
		count++
	}
	if count > config.MaxLengthUsedList {
		borderId := lastId - config.MaxLengthUsedList
		err = db.ExecTransact("DeletePasswords", borderId)
		if err != nil {
			log.Fatal(err)
		}
	}
}
