/*
Package webapp ...
	Runs webserver and displays content
*/
package webapp

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"main/webapp/controller"
	"main/webapp/middleware"
	"main/webapp/models"
	"main/webapp/protobuf"
	"main/webapp/service"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

func tempRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "webapp/templates/base.html", "webapp/templates/welcome.html")
	r.AddFromFiles("signup", "webapp/templates/base.html", "webapp/templates/signup.html")
	r.AddFromFiles("login", "webapp/templates/base.html", "webapp/templates/login.html")
	r.AddFromFiles("edit", "webapp/templates/base.html", "webapp/templates/edit.html")
	r.AddFromFiles("notfound", "webapp/templates/base.html", "webapp/templates/notfound.html")
	r.AddFromFiles("pantry", "webapp/templates/base.html", "webapp/templates/pantry.html")
	r.AddFromFiles("additem", "webapp/templates/base.html", "webapp/templates/additem.html")
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

	// NOTE : Get index page
	router.GET("/", func(c *gin.Context) {
		// Serve index.html
		message := authuser(c)
		if message != nil {
			c.HTML(200, "index", gin.H{"userobj": message.Claims.(jwt.MapClaims)["name"]})
			return
		}
		c.HTML(200, "index", gin.H{})
	})

	// API call to determine if user is valid
	router.GET("/authuser", func(c *gin.Context) {
		// Check cookie value is set and if cookie corresponds to valid JWT
		message := authuser(c)
		c.ProtoBuf(200, message)
	})

	// Present not found page
	router.GET("/notfound/:type", func(c *gin.Context) {
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

	// NOTE : signup user logic
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
			user := models.User{Email: email, Password: newpass}
			models.DB.Create(&user)
			userinfo := models.UserInfo{Restirctions: []string{}, City: "", State: "", ID: int(user.ID)}
			models.DB.Create(&userinfo)
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/notfound/signup")
	})

	// NOTE : user login logic
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
		c.Redirect(http.StatusFound, "/notfound/login")
	})

	// NOTE : edit user info logic
	router.GET("/useredit", func(c *gin.Context) {
		// Generate token
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			userinfo := models.UserInfo{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Where("ID = ?", user.ID).First(&userinfo)
			fmt.Println(userinfo)
			c.HTML(200, "edit", gin.H{"userobj": user, "city": userinfo.City, "state": userinfo.State, "restrictions": userinfo.Restirctions})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : edit user info logic
	router.POST("/edit_user", func(c *gin.Context) {
		// Generate token
		message := authuser(c)
		if message != nil {
			password := c.PostForm("Password")
			city := c.PostForm("City")
			state := c.PostForm("State")
			restricitons := c.PostFormArray("restrictions")
			fmt.Println(restricitons)
			// Hash and update password
			data := []byte(password)
			hash := md5.Sum(data)
			password = hex.EncodeToString(hash[:])
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			userinfo := models.UserInfo{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Where("ID = ?", user.ID).First(&userinfo)
			user.Password = password
			userinfo.City = city
			userinfo.State = state
			userinfo.Restirctions = restricitons
			models.DB.Save(&user)
			models.DB.Save(&userinfo)
			// c.HTML(200, "edit", gin.H{"userobj": user, "city": userinfo.City, "state": userinfo.State, "restrictions": userinfo.Restirctions})
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// Logout user
	router.GET("/logout", func(c *gin.Context) {
		// delete token cookie and send home
		c.SetCookie("token", "", -1, "/", "", false, false)
		c.Redirect(http.StatusFound, "/")
	})

	// NOTE : digital pantry logic
	router.GET("/pantry", func(c *gin.Context) {
		// Generate token
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			pantry := []models.Ingredient{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Order("expiration asc").Find(&pantry, "uid = ?", user.ID)
			fmt.Println(pantry)
			c.HTML(200, "pantry", gin.H{"userobj": user, "pantry": pantry})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.GET("/additem", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			c.HTML(200, "additem", gin.H{})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.POST("/additem", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			date := c.PostForm("Expiration")
			if date != "" {
				if invalidDate(date) {
					date = ""
				}
			}
			name := strings.Title(strings.ToLower(c.PostForm("Name")))
			quantity := c.PostForm("Quantity")
			if quantity == "0" {
				quantity = "N/A"
			}
			volume := c.PostForm("Volume")
			if volume != "0" {
				volume += (" " + c.PostForm("VolumeUnits"))
			} else {
				volume = "N/A"
			}
			weight := c.PostForm("Weight")
			if weight != "0" {
				weight += (" " + c.PostForm("WeightUnits"))
			} else {
				weight = "N/A"
			}
			image := c.PostForm("Image")
			if image == "Default" {
				url := "https://api.spoonacular.com/food/search?query=" + name + "&offset=0&number=1&apiKey=" + os.Getenv("APIKEY")
				method := "GET"
				client := &http.Client{}
				req, err := http.NewRequest(method, url, nil)
				if err != nil {
					fmt.Println(err)
					return
				}
				res, err := client.Do(req)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				jsonString := string(body)
				fmt.Println(jsonString)
				numResults := gjson.Get(jsonString, "searchResults.5")
				if numResults.Get("totalResults").Int() == 1 {
					image = "https://spoonacular.com/cdn/ingredients_100x100/" + numResults.Get("results.0.image").String()
				} else {
					image = ""
				}
				models.DB.Create(&models.Ingredient{
					Name:       name,
					UID:        uint(message.Claims.(jwt.MapClaims)["UID"].(float64)),
					Quantity:   quantity,
					Volume:     volume,
					Weight:     weight,
					ImageLink:  image,
					Expiration: date,
				})
			}
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			c.Redirect(http.StatusFound, "/pantry")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.Run()
}

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
