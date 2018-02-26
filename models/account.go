package models

import "time"

type Account struct {
	Id                 int
	Name               string
	Password           []byte
	PasswordUpdateTime time.Time
}
