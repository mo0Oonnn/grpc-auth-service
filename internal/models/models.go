package models

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash []byte
	IsAdmin      bool
}

type App struct {
	ID     int
	Name   string
	Secret string
}
