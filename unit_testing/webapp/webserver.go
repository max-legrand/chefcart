/*
Package webapp ...
	Runs webserver and displays content
*/
package webapp

import (
	"errors"
	"fmt"
	"log"
	"main/webapp/controller"
	"main/webapp/middleware"
	"main/webapp/models"
	Tokens "main/webapp/protobuf"
	"main/webapp/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

// Layers base template with relevant files
func tempRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "webapp/templates/base.html", "webapp/templates/welcome.html")
	r.AddFromFiles("signup", "webapp/templates/base.html", "webapp/templates/signup.html")
	r.AddFromFiles("login", "webapp/templates/base.html", "webapp/templates/login.html")
	r.AddFromFiles("edit", "webapp/templates/base.html", "webapp/templates/edit.html")
	r.AddFromFiles("notfound", "webapp/templates/base.html", "webapp/templates/notfound.html")
	r.AddFromFiles("pantry", "webapp/templates/base.html", "webapp/templates/pantry.html")
	r.AddFromFiles("additem", "webapp/templates/base.html", "webapp/templates/additem.html")
	r.AddFromFiles("edititem", "webapp/templates/base.html", "webapp/templates/edititem.html")
	r.AddFromFiles("recipe", "webapp/templates/base.html", "webapp/templates/recipe.html")
	r.AddFromFiles("recipeResults", "webapp/templates/base.html", "webapp/templates/recipeResults.html")
	// r.AddFromFiles("about", "templates/base.html", "templates/about.html")
	// r.AddFromFilesFuncs("about", template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }}, "templates/base.html", "templates/about.html")
	return r
}

// AuthUserUnwrapped - Unwrapped version of AuthUser function to be used for testing purpouses
func AuthUserUnwrapped(in *Tokens.Token) (*Tokens.Token, error) {
	token, valid := middleware.ValidTokenGRPC(in)
	if valid {
		isUserAuthenticated, _ := service.LoginUser(token.Claims.(jwt.MapClaims)["name"].(string), token.Claims.(jwt.MapClaims)["pass"].(string))
		if isUserAuthenticated {
			return &Tokens.Token{Token: token.Claims.(jwt.MapClaims)["name"].(string)}, nil
		}

	}
	return &Tokens.Token{}, nil
}

// GetPantryUnwrapped - get pantry items for user without grpc wrapper
func GetPantryUnwrapped(in *Tokens.Token) (*Tokens.Pantry, error) {
	token, valid := middleware.ValidTokenGRPC(in)
	if valid {
		// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
		user := models.User{}
		pantry := []models.Ingredient{}
		models.DB.Where("email = ?", token.Claims.(jwt.MapClaims)["name"]).First(&user)
		models.DB.Order("expiration asc").Find(&pantry, "uid = ?", user.ID)
		// fmt.Println(pantry)
		result := Tokens.Pantry{Pantry: []*Tokens.Ingredient{}}

		for _, item := range pantry {
			result.Pantry = append(result.Pantry, &Tokens.Ingredient{
				Id:         int64(item.ID),
				UID:        uint64(item.UID),
				Name:       item.Name,
				Quantity:   item.Quantity,
				Weight:     item.Weight,
				Volume:     item.Volume,
				Expiration: item.Expiration,
				ImageLink:  item.ImageLink,
			})
		}
		return &result, nil
	}
	return &Tokens.Pantry{}, errors.New("Invalid user")
}

// Determine if date is in a valid format
func invalidDate(dateString string) bool {
	monthsplit := strings.Index(dateString, "/")
	month := dateString[0:monthsplit]
	_, err := strconv.Atoi(month)
	if err != nil {
		return true
	}
	dateString = dateString[monthsplit+1:]
	daysplit := strings.Index(dateString, "/")
	day := dateString[0:daysplit]
	_, err = strconv.Atoi(day)
	if err != nil {
		return true
	}
	dateString = dateString[daysplit+1:]
	_, err = strconv.Atoi(dateString)
	return err != nil
}

//GetLoginToken - Get the login token from a gin context and verify its integrity
func GetLoginToken(loginController controller.LoginController, c *gin.Context) string {
	token := loginController.Login(c)
	if token != "" {
		encToken := middleware.Encrypt(token)
		// Set token to cookie & send back home
		message := &Tokens.Token{Token: encToken}
		data, err := proto.Marshal(message)
		stringarray := fmt.Sprint(data)
		stringarray = stringarray[1 : len(stringarray)-1]
		fmt.Println(stringarray)
		if err != nil {
			log.Fatal("marshaling error: ", err)
		}

		c.SetCookie("token", stringarray, 60*60*24, "/", "", false, false)
		c.Redirect(http.StatusFound, "/")
	}
	return token
}

// GetLoginTokenUnwrapped - login token function for use outside of gin webserver
func GetLoginTokenUnwrapped(loginController controller.LoginController, username, password *Tokens.Token) string {
	token := loginController.LoginTest(username, password)
	if token != "" {
		encToken := middleware.Encrypt(token)
		// Set token to cookie & send back home
		message := &Tokens.Token{Token: encToken}
		data, err := proto.Marshal(message)
		stringarray := fmt.Sprint(data)
		stringarray = stringarray[1 : len(stringarray)-1]
		// fmt.Println(stringarray)
		if err != nil {
			log.Fatal("marshaling error: ", err)
		}
		token = stringarray
	}
	return token
}

// Determine if a user is properly authenticated
func authuser(c *gin.Context) *jwt.Token {
	token, valid := middleware.ValidToken(c)
	// If valid send username from JWT
	if valid {
		isUserAuthenticated, _ := service.LoginUser(token.Claims.(jwt.MapClaims)["name"].(string), token.Claims.(jwt.MapClaims)["pass"].(string))
		if isUserAuthenticated {
			// message := &protobuf.Token{Token: token.Claims.(jwt.MapClaims)["name"].(string)}
			// data, _ := proto.Marshal(message)
			// stringarray := fmt.Sprint(data)
			// stringarray = stringarray[1 : len(stringarray)-1]
			return token
		}

	}
	// If not, send empty string
	// message := &protobuf.Token{Token: ""}
	return nil
}
