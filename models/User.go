package models


type User struct {
	Nickname     string `gorm:"primary_key"`
	FirstName    string
	LastName     string
	PasswordHash []byte
}
