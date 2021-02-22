/*
Package controller ...
	Gets login data from form and binds to Go Struct
*/
package controller

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"main/webapp/service"

	"github.com/gin-gonic/gin"
)

// LoginCredentials ...
type LoginCredentials struct {
	Email    string `form:"Email"`
	Password string `form:"Password"`
}

// LoginController ...
type LoginController interface {
	Login(ctx *gin.Context) string
}

type loginController struct {
	jWtService service.JWTService
}

// LoginHandler ...
func LoginHandler(jWtService service.JWTService) LoginController {
	return &loginController{
		jWtService: jWtService,
	}
}

func (controller *loginController) Login(ctx *gin.Context) string {
	// Gets login form information and binds to struct
	var credential LoginCredentials
	err := ctx.ShouldBind(&credential)
	fmt.Println(credential)
	if err != nil {
		return ""
	}
	// Hash and update password
	data := []byte(credential.Password)
	hash := md5.Sum(data)
	credential.Password = hex.EncodeToString(hash[:])

	// if user is valid, generate token
	isUserAuthenticated := service.LoginUser(credential.Email, credential.Password)
	if isUserAuthenticated {
		return controller.jWtService.GenerateToken(credential.Email, credential.Password, true)
	}
	return ""
}
