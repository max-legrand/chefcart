/*
Package webapp ...
	Runs webserver and displays content
*/
package webapp

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"main/webapp/controller"
	"main/webapp/middleware"
	"main/webapp/models"
	Tokens "main/webapp/protobuf"
	"main/webapp/service"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	mobile "github.com/floresj/go-contrib-mobile"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	ginsession "github.com/go-session/gin-session"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
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
	r.AddFromFiles("edititem", "webapp/templates/base.html", "webapp/templates/edititem.html")
	r.AddFromFiles("recipe", "webapp/templates/base.html", "webapp/templates/recipe.html")
	r.AddFromFiles("recipeResults", "webapp/templates/base.html", "webapp/templates/recipeResults.html")
	// r.AddFromFiles("about", "templates/base.html", "templates/about.html")
	// r.AddFromFilesFuncs("about", template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }}, "templates/base.html", "templates/about.html")
	return r
}

// Server ...
type Server struct {
	Tokens.UnimplementedServerServer
}

// NewServer ...
func NewServer() *Server {
	s := &Server{}
	return s
}

// AuthUser ...
func (c *Server) AuthUser(ctx context.Context, in *Tokens.Token) (*Tokens.Token, error) {
	token, valid := middleware.ValidTokenGRPC(in)
	if valid {
		isUserAuthenticated, _ := service.LoginUser(token.Claims.(jwt.MapClaims)["name"].(string), token.Claims.(jwt.MapClaims)["pass"].(string))
		if isUserAuthenticated {
			return &Tokens.Token{Token: token.Claims.(jwt.MapClaims)["name"].(string)}, nil
		}

	}
	return &Tokens.Token{}, nil
}

// GetPantry ...
func (c *Server) GetPantry(ctx context.Context, in *Tokens.Token) (*Tokens.Pantry, error) {
	token, valid := middleware.ValidTokenGRPC(in)
	if valid {
		// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
		user := models.User{}
		pantry := []models.Ingredient{}
		models.DB.Where("email = ?", token.Claims.(jwt.MapClaims)["name"]).First(&user)
		models.DB.Order("expiration asc").Find(&pantry, "uid = ?", user.ID)
		fmt.Println(pantry)
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
	return &Tokens.Pantry{}, nil
}

// GetUserInfo ...
func (c *Server) GetUserInfo(ctx context.Context, in *Tokens.Token) (*Tokens.UserInfo, error) {
	token, valid := middleware.ValidTokenGRPC(in)
	if valid {
		user := models.User{}
		userinfo := models.UserInfo{}
		models.DB.Where("email = ?", token.Claims.(jwt.MapClaims)["name"]).First(&user)
		models.DB.Where("ID = ?", user.ID).First(&userinfo)
		result := Tokens.UserInfo{
			City:              userinfo.City,
			State:             userinfo.State,
			Diets:             userinfo.Diets,
			Intolerances:      userinfo.Intolerances,
			QuantityThreshold: float32(userinfo.QuantityThreshold),
		}
		return &result, nil
	}
	return &Tokens.UserInfo{}, nil

}

// LaunchServer ...
func LaunchServer() {

	// JWT login setup
	jwtService := service.JWTAuthService()
	loginController := controller.LoginHandler(jwtService)
	s := grpc.NewServer()
	customServer := NewServer()
	Tokens.RegisterServerServer(s, customServer)
	wrappedGrpc := grpcweb.WrapServer(s, grpcweb.WithOriginFunc(func(origin string) bool {
		return true
	}))

	// Router & Template Setup
	router := gin.Default()
	router.Use(ginsession.New())
	router.HTMLRender = tempRender()
	router.Use(static.Serve("/js", static.LocalFile("webapp/templates/js", true)))
	router.Use(middleware.GinGrpcWebMiddleware(wrappedGrpc))
	router.Use(cors.New(cors.Config{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-GRPC-WEB"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           time.Duration(300) * time.Second,
	}))
	// Intiialize SQLite DB
	models.ConnectDB()
	router.Use(mobile.Resolver())

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
		d := mobile.GetDevice(c)
		isMobile := false
		if d.Mobile() {
			isMobile = true
		}
		c.HTML(200, "signup", gin.H{"isMobile": isMobile})
	})

	router.GET("/login", func(c *gin.Context) {
		d := mobile.GetDevice(c)
		isMobile := false
		if d.Mobile() {
			isMobile = true
		}
		c.HTML(200, "login", gin.H{"isMobile": isMobile})
	})

	// NOTE : signup user logic
	router.POST("/signup_user", func(c *gin.Context) {
		email := c.PostForm("Email")
		password := c.PostForm("Password")
		fmt.Println(email)
		data := []byte(password)
		hash := md5.Sum(data)
		newpass := hex.EncodeToString(hash[:])
		users := []models.User{}
		models.DB.Where("email = ?", email).Find(&users)
		if len(users) == 0 {
			user := models.User{Email: email, Password: newpass}
			models.DB.Create(&user)
			userinfo := models.UserInfo{QuantityThreshold: -1, Diets: []string{}, Intolerances: []string{}, City: "", State: "", ID: int(user.ID)}
			models.DB.Create(&userinfo)
			getLoginToken(loginController, c)
		}
		c.Redirect(http.StatusFound, "/notfound/signup")
	})

	// NOTE : user login logic
	router.POST("/login_user", func(c *gin.Context) {
		// Generate token
		token := getLoginToken(loginController, c)
		if token == "" {
			c.Redirect(http.StatusFound, "/notfound/login")
		}
	})

	// NOTE : edit user info logic
	router.GET("/useredit", func(c *gin.Context) {
		// Generate token
		d := mobile.GetDevice(c)
		isMobile := false
		if d.Mobile() {
			isMobile = true
		}
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			userinfo := models.UserInfo{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Where("ID = ?", user.ID).First(&userinfo)
			fmt.Println(userinfo)
			c.HTML(200, "edit", gin.H{"isMobile": isMobile, "userobj": user, "userinfo": userinfo, "city": userinfo.City, "state": userinfo.State, "diets": userinfo.Diets, "intolerances": userinfo.Intolerances})
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
			diets := c.PostFormArray("diets")
			intolerances := c.PostFormArray("intolerances")
			quantityThreshold, _ := strconv.ParseFloat(c.PostForm("QuantityThreshold"), 64)
			if quantityThreshold < 0 {
				quantityThreshold = -1.0
			}
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
			userinfo.Intolerances = intolerances
			userinfo.Diets = diets
			userinfo.QuantityThreshold = quantityThreshold
			models.DB.Save(&user)
			models.DB.Save(&userinfo)
			c.SetCookie("token", "", -1, "/", "", false, false)
			getLoginToken(loginController, c)
			return
			// c.HTML(200, "edit", gin.H{"userobj": user, "city": userinfo.City, "state": userinfo.State, "restrictions": userinfo.Restirctions})
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
			store := ginsession.FromContext(c)
			foodName, found := store.Get("invalidFood")
			if !found {
				foodName = ""
			}
			store.Delete("invalidFood")
			store.Save()
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			// user := models.User{}
			// pantry := []models.Ingredient{}
			// models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			// models.DB.Order("expiration asc").Find(&pantry, "uid = ?", user.ID)
			// fmt.Println(pantry)
			// c.HTML(200, "pantry", gin.H{"userobj": user, "pantry": pantry})
			c.HTML(200, "pantry", gin.H{"foodName": foodName.(string)})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.GET("/recipe", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			user := models.User{}
			userinfo := models.UserInfo{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Where("ID = ?", user.ID).First(&userinfo)
			fmt.Println(userinfo)
			c.HTML(200, "recipe", gin.H{"diets": userinfo.Diets, "intolerances": userinfo.Intolerances})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.POST("/recipeSearch", func(c *gin.Context) {

		type myForm struct {
			Ingredients           []string `form:"ingredients[]"`
			AdditionalIngredients string   `form:"additionalIngredients"`
			Diets                 []string `form:"diets"`
			Intolerances          []string `form:"intolerances"`
			Cuisine               []string `form:"cuisine"`
		}

		type recipe struct {
			Name      string
			ID        string
			Used      string
			Missing   string
			ImageLink string
		}

		formData := myForm{}
		c.ShouldBind(&formData)
		ingredients := strings.Join(formData.Ingredients, ",")
		if formData.AdditionalIngredients != "" && ingredients != "" {
			ingredients += "," + formData.AdditionalIngredients
		} else if formData.AdditionalIngredients != "" {
			ingredients += formData.AdditionalIngredients
		}
		offset := 0
		resultsSeen := 0

		url := "https://api.spoonacular.com/recipes/complexSearch?intolerances=" + strings.Join(formData.Intolerances, ",") + "&includeIngredients=" + ingredients + "&number=10&offset=" + strconv.Itoa(offset) + "&diet=" + strings.Join(formData.Diets, ",") + "&cuisine=" + strings.Join(formData.Cuisine, ",") + "&apiKey=" + os.Getenv("APIKEY")
		fmt.Println(url)
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
		totalResults := gjson.Get(jsonString, "totalResults").Int()
		if totalResults > 10 {
			totalResults = 10
		}

		results := []recipe{}

		for resultsSeen < int(totalResults) {
			for _, value := range gjson.Get(jsonString, "results").Array() {
				resultsSeen++
				results = append(results, recipe{
					Name:      value.Get("title").String(),
					ID:        value.Get("id").String(),
					Used:      value.Get("usedIngredientCount").String(),
					Missing:   value.Get("missedIngredientCount").String(),
					ImageLink: value.Get("image").String(),
				})
			}
			offset += 10
			if resultsSeen >= int(totalResults) {
				break
			}
			url := "https://api.spoonacular.com/recipes/complexSearch?intolerances=" + strings.Join(formData.Intolerances, ",") + "&includeIngredients=" + ingredients + "&number=10&offset=" + strconv.Itoa(offset) + "&diet=" + strings.Join(formData.Diets, ",") + "&cuisine=" + strings.Join(formData.Cuisine, ",") + "&apiKey=" + os.Getenv("APIKEY")
			method := "GET"
			fmt.Println(url)
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
			jsonString = string(body)
		}
		resultJSON, _ := json.Marshal(results)
		fmt.Println(len(results))
		c.HTML(http.StatusOK, "recipeResults", gin.H{"recipes": string(resultJSON)})
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

	// Note: Add item logic
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

			url := "https://spoonacular.com/api/tagFoods"
			method := "POST"

			payload := &bytes.Buffer{}
			writer := multipart.NewWriter(payload)
			_ = writer.WriteField("text", name)
			err := writer.Close()
			if err != nil {
				fmt.Println(err)
				return
			}

			client := &http.Client{}
			req, err := http.NewRequest(method, url, payload)

			if err != nil {
				fmt.Println(err)
				return
			}

			req.Header.Set("Content-Type", writer.FormDataContentType())
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
			fmt.Println(string(body))
			jsonString := string(body)

			if len(gjson.Get(jsonString, "annotations").Array()) == 0 {
				store := ginsession.FromContext(c)
				store.Set("invalidFood", name)
				store.Save()
				c.Redirect(http.StatusFound, "/pantry")
				return
			}

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
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			c.Redirect(http.StatusFound, "/pantry")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.GET("/edit/:id", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)

			user := models.User{}
			pantry := models.Ingredient{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			result := models.DB.Order("expiration asc").Find(&pantry, "uid = ? AND ID = ?", user.ID, c.Param("id")).First(&pantry)
			fmt.Println(pantry)
			if result.RowsAffected == 0 {
				c.Redirect(http.StatusFound, "/pantry")
				return
			}
			quantity := ""
			if pantry.Quantity == "N/A" {
				quantity = "0"
			} else {
				quantity = pantry.Quantity
			}

			weight := "0"
			weightUnits := "grams"
			if pantry.Weight == "N/A" {
				weight = "0"
			} else {
				stringList := strings.Split(pantry.Weight, " ")
				weight = stringList[0]
				weightUnits = stringList[1]
			}

			volume := "0"
			volumeUnits := "fl.oz"
			if pantry.Volume == "N/A" {
				volume = "0"
			} else {
				stringList := strings.Split(pantry.Volume, " ")
				volume = stringList[0]
				volumeUnits = stringList[1]
			}

			c.HTML(200, "edititem", gin.H{"userobj": user, "pantry": pantry, "quantity": quantity, "weight": weight, "weightUnits": weightUnits, "volume": volume, "volumeUnits": volumeUnits})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.GET("/delete/:id", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			pantry := models.Ingredient{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			result := models.DB.Order("expiration asc").Find(&pantry, "uid = ? AND ID = ?", user.ID, c.Param("id")).First(&pantry)
			fmt.Println(pantry)
			if result.RowsAffected == 0 {
				c.Redirect(http.StatusFound, "/pantry")
				return
			}
			models.DB.Delete(&pantry)
			c.Redirect(http.StatusFound, "/pantry")
			return
		}
		c.Redirect(http.StatusFound, "/pantry")
	})

	router.POST("/edit/:id", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			pantry := models.Ingredient{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			result := models.DB.Order("expiration asc").Find(&pantry, "uid = ? AND ID = ?", user.ID, c.Param("id")).First(&pantry)
			fmt.Println(pantry)
			if result.RowsAffected == 0 {
				c.Redirect(http.StatusFound, "/pantry")
				return
			}

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

			}
			pantry.Name = name
			pantry.Quantity = quantity
			pantry.Volume = volume
			pantry.Weight = weight
			pantry.ImageLink = image
			pantry.Expiration = date
			models.DB.Save(&pantry)

			c.Redirect(http.StatusFound, "/pantry")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.Run(":8080")
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

func getLoginToken(loginController controller.LoginController, c *gin.Context) string {
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
