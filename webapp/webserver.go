/*
Package webapp ...
	Runs webserver and displays content
*/
package webapp

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"main/webapp/controller"
	"main/webapp/middleware"
	"main/webapp/models"
	"main/webapp/protobuf"
	"main/webapp/service"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

func tempRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "Webapp/templates/base.html", "Webapp/templates/welcome.html")
	r.AddFromFiles("signup", "Webapp/templates/base.html", "Webapp/templates/signup.html")
	r.AddFromFiles("login", "Webapp/templates/base.html", "Webapp/templates/login.html")
	r.AddFromFiles("notfound", "Webapp/templates/base.html", "Webapp/templates/notfound.html")
	// r.AddFromFiles("about", "templates/base.html", "templates/about.html")
	// r.AddFromFilesFuncs("about", template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }}, "templates/base.html", "templates/about.html")
	return r
}

// LaunchServer ...
func LaunchServer() {

	// JWT login setup
	jwtService := service.JWTAuthService()
	loginController := controller.LoginHandler(jwtService)

	// Router & Template Setup
	router := gin.Default()
	router.HTMLRender = tempRender()
	router.Use(static.Serve("/js", static.LocalFile("Webapp/templates/js", true)))
	// Intiialize SQLite DB
	models.ConnectDB()

	// Get index page
	router.GET("/", func(c *gin.Context) {
		// Serve index.html
		c.HTML(200, "index", gin.H{})
	})

	// API call to determine if user is valid
	router.GET("/authuser", func(c *gin.Context) {
		// Check cookie value is set and if cookie corresponds to valid JWT
		token, valid := middleware.ValidToken(c)
		// If valid send username from JWT
		if valid {
			message := &protobuf.Token{Token: token.Claims.(jwt.MapClaims)["name"].(string)}
			data, _ := proto.Marshal(message)
			stringarray := fmt.Sprint(data)
			stringarray = stringarray[1 : len(stringarray)-1]
			c.ProtoBuf(200, message)

			return
		}
		// If not, send empty string
		message := &protobuf.Token{Token: ""}
		c.ProtoBuf(200, message)
	})

	// Present not found page
	router.GET("/notfound/", func(c *gin.Context) {
		// Get type url parameter
		// If param = "login" -> present invalid credentials, else present username already exists
		if c.Param("type") == "login" {
			c.HTML(200, "notfound", gin.H{"text": "Invalid credentials"})
		} else {
			c.HTML(200, "notfound", gin.H{"text": "User already exists"})
		}
	})

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(200, "signup", gin.H{})
	})
	router.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login", gin.H{})
	})

	router.POST("/signup_user", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")
		fmt.Println(email)
		data := []byte(password)
		hash := md5.Sum(data)
		newpass := hex.EncodeToString(hash[:])
		users := []models.User{}
		models.DB.Where("email = ?", email).Find(&users)
		if len(users) == 0 {
			models.DB.Create(&models.User{Email: email, Password: newpass})
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/notfound")
	})

	// Process user login
	router.POST("/login_user", func(c *gin.Context) {
		// Generate token
		token := loginController.Login(c)
		if token != "" {
			encToken := middleware.Encrypt(token)
			// Set token to cookie & send back home
			message := &protobuf.Token{Token: encToken}
			data, err := proto.Marshal(message)
			stringarray := fmt.Sprint(data)
			stringarray = stringarray[1 : len(stringarray)-1]
			fmt.Println(stringarray)
			if err != nil {
				log.Fatal("marshaling error: ", err)
			}

			c.SetCookie("token", stringarray, 48*60, "/", "", false, false)
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/notfound?")
	})

	// Logout user
	router.GET("/logout", func(c *gin.Context) {
		// delete token cookie and send home
		c.SetCookie("token", "", -1, "/", "", false, false)
		c.Redirect(http.StatusFound, "/")
	})

	router.Run()
}
