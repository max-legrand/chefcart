/*
Package service ...
	Verifies login data and determeines if user exists in database
*/
package service

import (
	"main/webapp/models"
)

// LoginService ...
type LoginService interface {
	LoginUser(email string, password string) bool
}

// type loginInformation struct {
// 	email    string
// 	password string
// }

// LoginUser ...
func LoginUser(email string, password string) (bool, uint) {
	// Check if user exists in DB
	users := []models.User{}
	models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
	if len(users) == 1 {
		return true, users[0].ID
	}
	return false, 0
}
