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
	"errors"
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
	r.AddFromFiles("grocery", "webapp/templates/base.html", "webapp/templates/grocery.html")
	r.AddFromFiles("addGrocery", "webapp/templates/base.html", "webapp/templates/addGrocery.html")
	r.AddFromFiles("groceryResults", "webapp/templates/base.html", "webapp/templates/groceryResults.html")
	r.AddFromFiles("changePassword", "webapp/templates/base.html", "webapp/templates/changePassword.html")
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

// AuthUser - verify user is valid from encrypted jtw token
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

// GetPantry - get pantry items for user
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
				Id:                int64(item.ID),
				UID:               uint64(item.UID),
				Name:              item.Name,
				Quantity:          item.Quantity,
				Weight:            item.Weight,
				Volume:            item.Volume,
				Expiration:        item.Expiration,
				ImageLink:         item.ImageLink,
				QuantityThreshold: float32(item.QuantityThreshold),
			})
		}
		return &result, nil
	}
	return &Tokens.Pantry{}, nil
}

// GetGroceries ...
func (c *Server) GetGroceries(ctx context.Context, in *Tokens.Token) (*Tokens.Pantry, error) {
	token, valid := middleware.ValidTokenGRPC(in)
	if valid {
		// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
		user := models.User{}
		pantry := []models.Grocery{}
		models.DB.Where("email = ?", token.Claims.(jwt.MapClaims)["name"]).First(&user)
		models.DB.Find(&pantry, "uid = ?", user.ID)
		fmt.Println(pantry)
		result := Tokens.Pantry{Pantry: []*Tokens.Ingredient{}}

		for _, item := range pantry {
			result.Pantry = append(result.Pantry, &Tokens.Ingredient{
				Id:        int64(item.ID),
				UID:       uint64(item.UID),
				Name:      item.Name,
				ImageLink: item.ImageLink,
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
			City:         userinfo.City,
			State:        userinfo.State,
			Diets:        userinfo.Diets,
			Intolerances: userinfo.Intolerances,
		}
		return &result, nil
	}
	return &Tokens.UserInfo{}, nil
}

// GetSearchResults ...
func (c *Server) GetSearchResults(ctx context.Context, in *Tokens.SearchQuery) (*Tokens.Store, error) {
	tokenStruct := &Tokens.Token{Token: in.Token}
	token, valid := middleware.ValidTokenGRPC(tokenStruct)
	if valid {
		// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
		user := models.User{}
		userinfo := models.UserInfo{}
		pantry := models.Grocery{}
		models.DB.Where("email = ?", token.Claims.(jwt.MapClaims)["name"]).First(&user)
		models.DB.Where("ID = ?", user.ID).First(&userinfo)
		result := models.DB.Find(&pantry, "uid = ? AND ID = ?", user.ID, in.ID).First(&pantry)
		fmt.Println(pantry)
		if result.RowsAffected == 0 {
			return &Tokens.Store{}, errors.New("Item does not belong to you")
		}

		url := "http://api.geonames.org/postalCodeSearchJSON?username=malaow3&placename=" + strings.ReplaceAll(userinfo.City, " ", "%20") + ",%20" + userinfo.State + "&placename_startsWith=" + strings.ReplaceAll(userinfo.City, " ", "%20") + "&maxRows=1&countryBias=US"
		fmt.Println(url)
		postalJSON := getFromURL(url, "")
		if len(gjson.Get(postalJSON, "postalCodes").Array()) == 0 {
			return &Tokens.Store{}, errors.New("No Stores Found")
		}
		postalCode := gjson.Get(postalJSON, "postalCodes").Array()[0].Get("postalCode").String()

		url = "https://www.walmart.com/grocery/v4/api/serviceAvailability?postalCode=" + postalCode
		walmartStores := getFromURL(url, "store")

		stores := gjson.Get(walmartStores, "accessPointList").Array()
		var myStore gjson.Result
		if len(stores) == 0 {
			return &Tokens.Store{}, errors.New("No Stores Found")
		}
		for _, store := range stores {
			if strings.HasPrefix(store.Get("name").String(), "Walmart") {
				myStore = store
				break
			}
		}
		storeResult := Tokens.Store{}
		storeResult.Address = myStore.Get("address.line1").String() + ", " + myStore.Get("address.city").String() + ", " + myStore.Get("address.state").String() + ", " + myStore.Get("address.postalCode").String()[0:5]
		storeResult.Monday = myStore.Get("workHours.monday").String()
		storeResult.Tuesday = myStore.Get("workHours.tuesday").String()
		storeResult.Wednesday = myStore.Get("workHours.wednesday").String()
		storeResult.Thursday = myStore.Get("workHours.thursday").String()
		storeResult.Friday = myStore.Get("workHours.friday").String()
		storeResult.Saturday = myStore.Get("workHours.saturday").String()
		storeResult.Sunday = myStore.Get("workHours.sunday").String()
		storeResult.Distance = myStore.Get("distance").String()
		storeID := myStore.Get("dispenseStoreId").String()
		url = "https://www.walmart.com/grocery/v4/api/products/search?count=25&offset=0&page=1&storeId=" + storeID + "&query=" + pantry.Name
		foodJSON := getFromURL(url, "food")
		foodArray := gjson.Get(foodJSON, "products").Array()
		foodResults := []*Tokens.Food{}
		for _, foodItem := range foodArray {
			foodStruct := Tokens.Food{}
			foodStruct.Name = foodItem.Get("basic.name").String()
			foodStruct.Image = foodItem.Get("basic.image.thumbnail").String()
			foodStruct.Link = "https://www.walmart.com" + foodItem.Get("basic.productUrl").String()
			foodStruct.Rating = float32(foodItem.Get("detailed.rating").Float())
			foodStruct.Reviews = foodItem.Get("detailed.reviewsCount").Int()
			foodStruct.InStock = !foodItem.Get("store.isOutOfStock").Bool()
			foodStruct.Price = float32(foodItem.Get("store.price.displayPrice").Float())
			foodResults = append(foodResults, &foodStruct)
		}
		storeResult.Results = foodResults
		return &storeResult, nil

	}
	return &Tokens.Store{}, nil

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

	// NOTE : Present not found page
	router.GET("/notfound/:type", func(c *gin.Context) {
		// Get type url parameter
		// If param = "login" -> present invalid credentials, else present username already exists
		if c.Param("type") == "login" {
			c.HTML(200, "notfound", gin.H{"text": "Invalid credentials"})
		} else {
			c.HTML(200, "notfound", gin.H{"text": "User already exists"})
		}
	})

	// NOTE : Present the signup form
	router.GET("/signup", func(c *gin.Context) {
		d := mobile.GetDevice(c)
		isMobile := false
		if d.Mobile() {
			isMobile = true
		}
		c.HTML(200, "signup", gin.H{"isMobile": isMobile})
	})

	// NOTE : Present the login form
	router.GET("/login", func(c *gin.Context) {
		d := mobile.GetDevice(c)
		isMobile := false
		if d.Mobile() {
			isMobile = true
		}
		c.HTML(200, "login", gin.H{"isMobile": isMobile})
	})

	// NOTE : Perform signup user logic
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
			userinfo := models.UserInfo{Diets: []string{}, Intolerances: []string{}, City: "", State: "", ID: int(user.ID)}
			models.DB.Create(&userinfo)
			getLoginToken(loginController, c)
			return
		}
		c.Redirect(http.StatusFound, "/notfound/signup")
	})

	// NOTE : Perform user login logic
	router.POST("/login_user", func(c *gin.Context) {
		// Generate token
		token := getLoginToken(loginController, c)
		if token == "" {
			c.Redirect(http.StatusFound, "/notfound/login")
		}
	})

	// NOTE : Display user edit form
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

	// NOTE : Perform edit user info logic
	router.POST("/edit_user", func(c *gin.Context) {
		// Generate token
		message := authuser(c)
		if message != nil {
			city := c.PostForm("City")
			state := c.PostForm("State")
			diets := c.PostFormArray("diets")
			intolerances := c.PostFormArray("intolerances")
			quantityThreshold, _ := strconv.ParseFloat(c.PostForm("QuantityThreshold"), 64)
			if quantityThreshold < 0 {
				quantityThreshold = -1.0
			}
			user := models.User{}
			userinfo := models.UserInfo{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Where("ID = ?", user.ID).First(&userinfo)
			userinfo.City = city
			userinfo.State = state
			userinfo.Intolerances = intolerances
			userinfo.Diets = diets
			models.DB.Save(&user)
			models.DB.Save(&userinfo)
			c.SetCookie("token", "", -1, "/", "", false, false)
			getLoginToken(loginController, c)
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Logout user
	router.GET("/logout", func(c *gin.Context) {
		// delete token cookie and send home
		c.SetCookie("token", "", -1, "/", "", false, false)
		c.Redirect(http.StatusFound, "/")
	})

	// NOTE : Digital pantry logic
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
			c.HTML(200, "pantry", gin.H{"userobj": message.Claims.(jwt.MapClaims)["name"], "foodName": foodName.(string)})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Display recipe request form
	router.GET("/recipe", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			user := models.User{}
			userinfo := models.UserInfo{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Where("ID = ?", user.ID).First(&userinfo)
			fmt.Println(userinfo)
			c.HTML(200, "recipe", gin.H{"userobj": message.Claims.(jwt.MapClaims)["name"], "diets": userinfo.Diets, "intolerances": userinfo.Intolerances})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Perform recipe search
	router.POST("/recipeSearch", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			type myForm struct {
				Ingredients           []string `form:"ingredients[]"`
				AdditionalIngredients string   `form:"additionalIngredients"`
				ExcludedIngredients   string   `form:"excludedIngredients"`
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

			url := "https://api.spoonacular.com/recipes/complexSearch?intolerances=" + strings.Join(formData.Intolerances, ",") + "&includeIngredients=" + ingredients + "&excludeIngredients=" + formData.ExcludedIngredients + "&number=10&offset=" + strconv.Itoa(offset) + "&diet=" + strings.Join(formData.Diets, ",") + "&cuisine=" + strings.Join(formData.Cuisine, ",") + "&apiKey=" + os.Getenv("APIKEY")
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
			c.HTML(http.StatusOK, "recipeResults", gin.H{"userobj": message.Claims.(jwt.MapClaims)["name"], "recipes": string(resultJSON)})
			return
		}
		c.Redirect(http.StatusFound, "/login")

	})

	// NOTE : Display add item form
	router.GET("/additem", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			c.HTML(200, "additem", gin.H{"userobj": message.Claims.(jwt.MapClaims)["name"]})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Add item logic
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
			foodItem := models.Ingredient{}
			user := models.User{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)

			result := models.DB.Where("name = ? and UID = ?", name, user.ID).Find(&foodItem)
			if result.RowsAffected != 0 {
				store := ginsession.FromContext(c)
				if name[len(name)-1] == 's' {
					store.Set("invalidFood", name+" already exist in your digital pantry")
				} else {
					store.Set("invalidFood", name+" already exists in your digital pantry")
				}
				store.Save()
				c.Redirect(http.StatusFound, "/pantry")
				return
			}
			if result.RowsAffected != 0 {
				store := ginsession.FromContext(c)
				if name[len(name)-1] == 's' {
					store.Set("invalidFood", name+" already exist in your digital pantry")
				} else {
					store.Set("invalidFood", name+" already exists in your digital pantry")
				}
				store.Save()
				c.Redirect(http.StatusFound, "/pantry")
				return
			}
			if name[len(name)-1] == 's' {
				result := models.DB.Where("name = ? and UID = ?", name[0:len(name)-1], user.ID).Find(&foodItem)
				if result.RowsAffected != 0 {
					store := ginsession.FromContext(c)
					if name[len(name)-1] == 's' {
						store.Set("invalidFood", name+" already exist in your digital pantry")
					} else {
						store.Set("invalidFood", name+" already exists in your digital pantry")
					}
					store.Save()
					c.Redirect(http.StatusFound, "/pantry")
					return
				}
			} else {
				result := models.DB.Where("name = ? and UID = ?", name+"s", user.ID).Find(&foodItem)
				if result.RowsAffected != 0 {
					store := ginsession.FromContext(c)
					if name[len(name)-1] == 's' {
						store.Set("invalidFood", name+" already exist in your digital pantry")
					} else {
						store.Set("invalidFood", name+" already exists in your digital pantry")
					}
					store.Save()
					c.Redirect(http.StatusFound, "/pantry")
					return
				}
			}
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

			image := c.PostForm("Image")
			if len(gjson.Get(jsonString, "annotations").Array()) == 0 {
				store := ginsession.FromContext(c)
				store.Set("invalidFood", name+" is not a valid food item")
				store.Save()
				c.Redirect(http.StatusFound, "/pantry")
				return
			}
			if image == "Default" {
				image = gjson.Get(jsonString, "annotations.0.image").String()
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
			quantityThreshold, _ := strconv.ParseFloat(c.PostForm("QuantityThreshold"), 64)
			if quantityThreshold < 0 {
				quantityThreshold = -1
			}

			models.DB.Create(&models.Ingredient{
				Name:              name,
				UID:               uint(message.Claims.(jwt.MapClaims)["UID"].(float64)),
				Quantity:          quantity,
				Volume:            volume,
				Weight:            weight,
				ImageLink:         image,
				Expiration:        date,
				QuantityThreshold: quantityThreshold,
			})
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			c.Redirect(http.StatusFound, "/pantry")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Display edit pantry item form
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

	// NOTE : Delete pantry item
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

	// NOTE : Perform item edit
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

			image := c.PostForm("Image")
			if len(gjson.Get(jsonString, "annotations").Array()) == 0 {
				store := ginsession.FromContext(c)
				store.Set("invalidFood", name+" is not a valid food item")
				store.Save()
				c.Redirect(http.StatusFound, "/pantry")
				return
			}
			if image == "Default" {
				image = gjson.Get(jsonString, "annotations.0.image").String()
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
			quantityThreshold, _ := strconv.ParseFloat(c.PostForm("QuantityThreshold"), 64)
			if quantityThreshold < 0 {
				quantityThreshold = -1
			}
			pantry.Name = name
			pantry.Quantity = quantity
			pantry.Volume = volume
			pantry.Weight = weight
			pantry.ImageLink = image
			pantry.Expiration = date
			pantry.QuantityThreshold = quantityThreshold
			models.DB.Save(&pantry)

			c.Redirect(http.StatusFound, "/pantry")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Display Grocery list items
	router.GET("/grocery", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			store := ginsession.FromContext(c)
			foodName, found := store.Get("invalidFood")
			if !found {
				foodName = ""
			}
			store.Delete("invalidFood")
			store.Save()
			c.HTML(200, "grocery", gin.H{"userobj": message.Claims.(jwt.MapClaims)["name"], "foodName": foodName.(string)})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Add grocery list item
	router.GET("/addGrocery", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			c.HTML(200, "addGrocery", gin.H{"userobj": message.Claims.(jwt.MapClaims)["name"]})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Add grocery list item logic
	router.POST("/addGrocery", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {

			foodItem := models.Ingredient{}
			user := models.User{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			name := strings.Title(strings.ToLower(c.PostForm("Name")))

			result := models.DB.Where("name = ? and UID = ?", name, user.ID).Find(&foodItem)
			if result.RowsAffected != 0 {
				store := ginsession.FromContext(c)
				if name[len(name)-1] == 's' {
					store.Set("invalidFood", name+" already exist in your grocery list")
				} else {
					store.Set("invalidFood", name+" already exists in your grocery list")

				}
				store.Save()
				c.Redirect(http.StatusFound, "/grocery")
				return
			}
			if name[len(name)-1] == 's' {
				result := models.DB.Where("name = ? and UID = ?", name[0:len(name)-1], user.ID).Find(&foodItem)
				if result.RowsAffected != 0 {
					store := ginsession.FromContext(c)
					if name[len(name)-1] == 's' {
						store.Set("invalidFood", name+" already exist in your grocery list")
					} else {
						store.Set("invalidFood", name+" already exists in your grocery list")

					}
					store.Save()
					c.Redirect(http.StatusFound, "/grocery")
					return
				}
			} else {
				result := models.DB.Where("name = ? and UID = ?", name+"s", user.ID).Find(&foodItem)
				if result.RowsAffected != 0 {
					store := ginsession.FromContext(c)
					if name[len(name)-1] == 's' {
						store.Set("invalidFood", name+" already exist in your grocery list")
					} else {
						store.Set("invalidFood", name+" already exists in your grocery list")

					}
					store.Save()
					c.Redirect(http.StatusFound, "/grocery")
					return
				}
			}

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

			image := c.PostForm("Image")
			if len(gjson.Get(jsonString, "annotations").Array()) == 0 {
				store := ginsession.FromContext(c)
				store.Set("invalidFood", name+" is not a valid food")
				store.Save()
				c.Redirect(http.StatusFound, "/grocery")
				return
			}
			if image == "Default" {
				image = gjson.Get(jsonString, "annotations.0.image").String()
			}
			models.DB.Create(&models.Grocery{
				Name:      name,
				UID:       uint(message.Claims.(jwt.MapClaims)["UID"].(float64)),
				ImageLink: image,
			})
			c.Redirect(http.StatusFound, "/grocery")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Delete grocery list item
	router.GET("/deleteGrocery/:id", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			pantry := models.Grocery{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			result := models.DB.Find(&pantry, "uid = ? AND ID = ?", user.ID, c.Param("id")).First(&pantry)
			fmt.Println(pantry)
			if result.RowsAffected == 0 {
				c.Redirect(http.StatusFound, "/grocery")
				return
			}
			models.DB.Delete(&pantry)
			c.Redirect(http.StatusFound, "/grocery")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Search for grocery list item
	router.GET("/search/:id", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			userinfo := models.UserInfo{}
			pantry := models.Grocery{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			models.DB.Where("ID = ?", user.ID).First(&userinfo)
			result := models.DB.Find(&pantry, "uid = ? AND ID = ?", user.ID, c.Param("id")).First(&pantry)
			fmt.Println(pantry)
			if result.RowsAffected == 0 {
				c.Redirect(http.StatusFound, "/grocery")
				return
			}
			c.HTML(http.StatusOK, "groceryResults", gin.H{"userobj": user, "id": c.Param("id")})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// NOTE : Edit password
	router.GET("/changePassword", func(c *gin.Context) {
		d := mobile.GetDevice(c)
		isMobile := false
		if d.Mobile() {
			isMobile = true
		}
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			user := models.User{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			c.HTML(200, "changePassword", gin.H{"isMobile": isMobile, "userobj": user})
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.POST("/changePassword", func(c *gin.Context) {
		message := authuser(c)
		if message != nil {
			// models.DB.Where("email = ? AND password = ?", email, password).Find(&users)
			// Hash and update password
			password := c.PostForm("oldPass")
			data := []byte(password)
			hash := md5.Sum(data)
			password = hex.EncodeToString(hash[:])
			user := models.User{}
			models.DB.Where("email = ?", message.Claims.(jwt.MapClaims)["name"]).First(&user)
			if user.Password != password {
				c.Redirect(http.StatusFound, "/notfound/login")
				return
			}
			newPass := c.PostForm("Password")
			data = []byte(newPass)
			hash = md5.Sum(data)
			newpassword := hex.EncodeToString(hash[:])
			user.Password = newpassword
			models.DB.Save(&user)
			c.SetCookie("token", "", -1, "/", "", false, false)
			getLoginToken(loginController, c)
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/login")
	})

	router.Run()
}

func getFromURL(url string, walmart string) string {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	if walmart == "food" {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Accept-Language", "en-us")
		// req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		req.Header.Add("Host", "www.walmart.com")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15")
		req.Header.Add("Referer", "https://www.walmart.com/grocery/search/?query=pizza")
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Cookie", "TB_SFOU-100=; TS013ed49a=01538efd7cf6a0e1e3467f211498e9938b07ccb7019fc2f4e92fbfc51e9feebe2175615f48a9d5d7276a1a28dac22b6e74f7c4e653; bstc=SvIK5kMsYLTEqtq0yE3j7Y; exp-ck=Aa-Hd1coPFn1q7FRT1; mobileweb=0; vtc=S51vv5NXKeZVTw7VVq68Mc; xpa=1wERM|4ZRB4|5KNdl|Aa-Hd|coPFn|q7FRT|qnjoe; xpm=1%2B1616562069%2BS51vv5NXKeZVTw7VVq68Mc~dabc72e4-8920-424d-be60-842367422ceb%2B0; TS01b0be75=01538efd7c373f923e7a024dcfa8a18e7fb88c0d4754377d5009adfe78caaf0ba63fcf1fc3eadf7a58648a3d2966a1b2832cef4a31; akavpau_p14=1616564807~id=213d482646baae28fe355028f1ed371e; _pxde=54f2e895f871407c315d43407f6a54f31f8092a4c56c5535ace272426fc89f5d:eyJ0aW1lc3RhbXAiOjE2MTY1NjQyMDU0NDAsImZfa2IiOjAsImlwY19pZCI6W119; ACID=ea5ba3d0-8c62-11eb-8eac-ad5b2ad02634; GCRT=526de4dc-f2f1-405a-9491-263ba68f8097; TS012c809b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; TS01af768b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; hasACID=1; hasGCRT=1; wm_mystore=Fe26.2**232a0f87112ab8b97b05c89604142d25f5c69e313bc0496e6623af18ea224b98*wVsfFcU-eJ8RHWNAijsnlg*b4K_UN0v2rC4pbrRqWuha-vqQ25Fo8TclobrskCPU0FWfALZ6erW_gXAKvinDBJ7nkJEgRnCCEFo8TVCLuoxMigmx_xVjXjdKEJo1CWVIrIlsoup_7UstozDBLlgxIO5_4bFXedlrnoJJO8tLlCx5A**e6e63046e6e16ce85a05216d0033070c60da3cbd3c77cfd7e5f12e335f589ac6*5w_-IpnwNF5VQOu2fb8kzd0u8T4v14TsF2la9HFtG5Q; OG_CXO_CART_ST=%7B%22mtoken%22%3A%2241%3A5%23528694727%237%3D581686586%22%2C%22itoken%22%3A%7B%22UK_CART_OWNER_T_V%22%3A%2228%3A4%23101085325%237%3D103795201%22%7D%7D; TS01bae75b=01538efd7c1bc57bd268a06f6694a2a5c8561076901c24da6fbf5fa65fbc46b919c6d4c43dddd1d8ebb2722481669554c55bb4b66e; x-csrf-jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiY29va2llIiwidXVpZCI6ImViMzlhN2MwLThjNjItMTFlYi05ZGY5LTY1NDFjOTBlOTRhYiIsImlhdCI6MTYxNjU2NDIwNCwiZXhwIjoxNjE3NjQ0MjA0fQ.OfegAUAosjz3yRtmqPzZ9l4ChFRRgNtJuazGfMubIaw; _px3=848b4e5ca33e918a713ceb18a7de8d3cbe738c43d78e6aacccf0056ef9e64224:L2keyT+HM+LpFZjL3du1XpgPaW/EANaOgtlPP3nNVLqijD7EOriFb2+Frc7sa56lLp9NTgj6e3HkhzvrgmqPZw==:1000:l8XYFaMi/E+GykMtMtZx6OHO6583Mw7av2YrrpfaoB9t3w5Rb+ZKaKQmirLLO2PG4XSw7Zd/T9cqa9jsG0VvFHRsTxLlUUGEFGtUaOWNKZ4xy9iRHEuQhMghvte4endhO0fKGp5LL6YQ5fOQlD6wHlXdO4Rhv4loV2NG/ZDZoIg=; cart-item-count=0; DL=08873%2C%2C%2Cip%2C08873%2C%2C; com.wm.reflector=\"reflectorid:0000000000000000000000@lastupd:1616564163940@firstcreate:1616549589548\"; next-day=1616634000|true|false|1616673600|1616564164; akavpau_p8=1616564764~id=af0f7a29c7ee9912d94938046b742956; auth=MTAyOTYyMDE4TxS3fbFODPleRTfmPyKSma%2BbeTUNULE7MaN2Ywm%2F%2F%2Bd5hYD91eGWPyuZoh7upNtLwP98Vmdia081SVPLqQpIzrcj2Qdw0VTxKGr0Yv%2FdDreABuZPdeU1b5hhEBS%2BJVOl4OnKIM5mq%2BMWt1pJmBJeTpytkL%2FujXG3kzuvFZppImS%2FUo8uubRCSkj%2BkcbVjGizJwoI882Ka5HFNGgeht2X4hQyTvQ%2FyCWA4Sks6muEN8zRdilwc3qC8sSv9s9HrEvt3aPZmWmFlT%2B7TQ%2B%2BVdpBxtRmJ7rvPnANTh5VWeInxo8uJ4K703bIF651A9DWMNXn1XLG27nQ9dkExVppnQ82bcvxjvWLwvtNUApOFsHuvpczHXpXOf%2FoXkEmE12JACOA; location-data=08873%3ASomerset%3ANJ%3A%3A1%3A1|1jn%3B%3B3.68%2C21n%3B%3B4.87%2C215%3B%3B6.03%2C2di%3B%3B6.94%2C40h%3B%3B8.15%2C3xz%3B%3B11.24%2C26h%3B%3B12.09%2C42p%3B%3B12.41%2C37d%3B%3B13.53%2C2iq%3B%3B16.62||7|1|1yl1%3B16%3B4%3B8.01%2C1yfy%3B16%3B11%3B17.73%2C1xgb%3B16%3B12%3B19.72%2C1ym1%3B16%3B13%3B24.41%2C1xn0%3B16%3B14%3B24.93; CID=dabc72e4-8920-424d-be60-842367422ceb; WMP=4; customer=%7B%22firstName%22%3A%22mala%22%2C%22lastNameInitial%22%3A%22o%22%2C%22rememberme%22%3Atrue%7D; hasCID=1; oneapp_customer=true; rtoken=MDgyNTUyMDE48574DBWi3V0Qv25koDr2HXVqsf%2BFQEzRwNRiS1YkJiRj8pXP0eTNRdC8umTl53NUVRi8xXvNG99kaWygC0TpFXDXf%2FdcPfsPS8io4VyezDmoudDRc7ekLUDT%2BOCGtMbK5HlODFAHE8TT0qFXMh%2BBLJaUrJ%2BFHjmcN%2B7tjMwZbr7FUjvszEqIm71dht90KFkhvXRZeR3z4CP5F0qVH0j%2Fnjrj2Zlfg7NLwazYzfRa78HBK95JK6Di2imC69hk27N%2BYII6CMcLJmM7tvaYHp90tpCcFWH1yoILr3SvZRFA8k%2BiZtwxikEy7pU44SSVVR027nZI65tkEiR4JN00qrkUVYMNm76NCRKaUZlizpp3TQJ8ghRlFDFOojQ66oL9rikxuiKqfgImRu%2BgI1PqX7lxWg%3D%3D; type=REGISTERED; wm_ul_plus=INACTIVE|1616648469715; ONEAPP_CUSTOMER=true; member=0; TB_DC_Flap_Test=0; athrvi=RVI~ha4d66df; _abck=hqwoigx515i3t5ob07av_1782; _pxvid=d728ca08-8c40-11eb-bf0e-0242ac120007; tb_sw_supported=false; TBV=13; TB_Latency_Tracker_100=1; TB_Navigation_Preload_01=1; TB_DC_Flap_Test=1; TB_Latency_Tracker_100=1; TB_Navigation_Preload_01=1; TB_SFOU-100=; TS012c809b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; TS01af768b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; bstc=SvIK5kMsYLTEqtq0yE3j7Y; exp-ck=44e3B14644E1Aa-Hd1vZriO1; g=0; mobileweb=0; vtc=S51vv5NXKeZVTw7VVq68Mc; xpa=44e3B|4644E|Aa-Hd|GLQ0p|vZriO; xpm=7%2B1616564245%2BbxpoIRrPw1pR5N4Jb6zHxo~dabc72e4-8920-424d-be60-842367422ceb%2B0; TS01bae75b=01538efd7c6b7b00aebf23763a1ddc5619c4e4dfe0276b30b5a16831a4d8c6ebdb5b4ed08af619d621174c92848b8b86206bb0855f; x-csrf-jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiY29va2llIiwidXVpZCI6IjMxYTEwODIwLThjNjMtMTFlYi1iN2I0LTExY2ZmMjc4YTc3MiIsImlhdCI6MTYxNjU2NDMyMiwiZXhwIjoxNjE3NjQ0MzIyfQ.3RhSZiU4CxC5vcjyAP8JfnIm-z8kfzP_CgsT_AWmQSY")
		req.Header.Add("x-csrf-jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiaGVhZGVyIiwidXVpZCI6ImViMzlhN2MwLThjNjItMTFlYi05ZGY5LTY1NDFjOTBlOTRhYiIsImlhdCI6MTYxNjU2NDIwNCwiZXhwIjozMjY2ODU2NTg5NX0.lBxXzxG5cOCmHyInz5s8jIn_tkbiEtn5R54JSR1gaVM")
		req.Header.Add("wm_vertical_id", "2")
		req.Header.Add("wg-correlation-id", "ed019680-8c62-11eb-ad85-290b60e46528")
		req.Header.Add("wm_tenant_id", "0")
	} else if walmart == "store" {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Cookie", "akavpau_p14=1616565068~id=311643d9745ebe20c6d24cb102cdecf2; TB_SFOU-100=; TS012c809b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; TS01af768b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; bstc=SvIK5kMsYLTEqtq0yE3j7Y; exp-ck=Aa-Hd1coPFn1q7FRT1; mobileweb=0; vtc=S51vv5NXKeZVTw7VVq68Mc; xpa=1wERM|4ZRB4|5KNdl|Aa-Hd|coPFn|q7FRT|qnjoe; xpm=1%2B1616562069%2BS51vv5NXKeZVTw7VVq68Mc~dabc72e4-8920-424d-be60-842367422ceb%2B0; TS01bae75b=01538efd7c1bc57bd268a06f6694a2a5c8561076901c24da6fbf5fa65fbc46b919c6d4c43dddd1d8ebb2722481669554c55bb4b66e; TB_DC_Flap_Test=0; _pxde=8231a9cc7a5a47d82cc7875ffe74ab009ab6b62dc4a96549285af255e249bc75:eyJ0aW1lc3RhbXAiOjE2MTY1NjQ0NjQxNjgsImZfa2IiOjAsImlwY19pZCI6W119; x-csrf-jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiY29va2llIiwidXVpZCI6ImVkMGJhOGEwLThjNjItMTFlYi1hMTkwLWRmYjk5MDhlNjg0YSIsImlhdCI6MTYxNjU2NDIwNywiZXhwIjoxNjE3NjQ0MjA3fQ.JwUameNcXgZT7a71K--o66CtH7qVPIMNPFZwZIY1fzI; TS013ed49a=01538efd7cf6a0e1e3467f211498e9938b07ccb7019fc2f4e92fbfc51e9feebe2175615f48a9d5d7276a1a28dac22b6e74f7c4e653; TS01b0be75=01538efd7c373f923e7a024dcfa8a18e7fb88c0d4754377d5009adfe78caaf0ba63fcf1fc3eadf7a58648a3d2966a1b2832cef4a31; ACID=ea5ba3d0-8c62-11eb-8eac-ad5b2ad02634; GCRT=526de4dc-f2f1-405a-9491-263ba68f8097; hasACID=1; hasGCRT=1; wm_mystore=Fe26.2**232a0f87112ab8b97b05c89604142d25f5c69e313bc0496e6623af18ea224b98*wVsfFcU-eJ8RHWNAijsnlg*b4K_UN0v2rC4pbrRqWuha-vqQ25Fo8TclobrskCPU0FWfALZ6erW_gXAKvinDBJ7nkJEgRnCCEFo8TVCLuoxMigmx_xVjXjdKEJo1CWVIrIlsoup_7UstozDBLlgxIO5_4bFXedlrnoJJO8tLlCx5A**e6e63046e6e16ce85a05216d0033070c60da3cbd3c77cfd7e5f12e335f589ac6*5w_-IpnwNF5VQOu2fb8kzd0u8T4v14TsF2la9HFtG5Q; _px3=848b4e5ca33e918a713ceb18a7de8d3cbe738c43d78e6aacccf0056ef9e64224:L2keyT+HM+LpFZjL3du1XpgPaW/EANaOgtlPP3nNVLqijD7EOriFb2+Frc7sa56lLp9NTgj6e3HkhzvrgmqPZw==:1000:l8XYFaMi/E+GykMtMtZx6OHO6583Mw7av2YrrpfaoB9t3w5Rb+ZKaKQmirLLO2PG4XSw7Zd/T9cqa9jsG0VvFHRsTxLlUUGEFGtUaOWNKZ4xy9iRHEuQhMghvte4endhO0fKGp5LL6YQ5fOQlD6wHlXdO4Rhv4loV2NG/ZDZoIg=; cart-item-count=0; DL=08873%2C%2C%2Cip%2C08873%2C%2C; com.wm.reflector=\"reflectorid:0000000000000000000000@lastupd:1616564163940@firstcreate:1616549589548\"; next-day=1616634000|true|false|1616673600|1616564164; akavpau_p8=1616564764~id=af0f7a29c7ee9912d94938046b742956; auth=MTAyOTYyMDE4TxS3fbFODPleRTfmPyKSma%2BbeTUNULE7MaN2Ywm%2F%2F%2Bd5hYD91eGWPyuZoh7upNtLwP98Vmdia081SVPLqQpIzrcj2Qdw0VTxKGr0Yv%2FdDreABuZPdeU1b5hhEBS%2BJVOl4OnKIM5mq%2BMWt1pJmBJeTpytkL%2FujXG3kzuvFZppImS%2FUo8uubRCSkj%2BkcbVjGizJwoI882Ka5HFNGgeht2X4hQyTvQ%2FyCWA4Sks6muEN8zRdilwc3qC8sSv9s9HrEvt3aPZmWmFlT%2B7TQ%2B%2BVdpBxtRmJ7rvPnANTh5VWeInxo8uJ4K703bIF651A9DWMNXn1XLG27nQ9dkExVppnQ82bcvxjvWLwvtNUApOFsHuvpczHXpXOf%2FoXkEmE12JACOA; location-data=08873%3ASomerset%3ANJ%3A%3A1%3A1|1jn%3B%3B3.68%2C21n%3B%3B4.87%2C215%3B%3B6.03%2C2di%3B%3B6.94%2C40h%3B%3B8.15%2C3xz%3B%3B11.24%2C26h%3B%3B12.09%2C42p%3B%3B12.41%2C37d%3B%3B13.53%2C2iq%3B%3B16.62||7|1|1yl1%3B16%3B4%3B8.01%2C1yfy%3B16%3B11%3B17.73%2C1xgb%3B16%3B12%3B19.72%2C1ym1%3B16%3B13%3B24.41%2C1xn0%3B16%3B14%3B24.93; CID=dabc72e4-8920-424d-be60-842367422ceb; WMP=4; customer=%7B%22firstName%22%3A%22mala%22%2C%22lastNameInitial%22%3A%22o%22%2C%22rememberme%22%3Atrue%7D; hasCID=1; oneapp_customer=true; rtoken=MDgyNTUyMDE48574DBWi3V0Qv25koDr2HXVqsf%2BFQEzRwNRiS1YkJiRj8pXP0eTNRdC8umTl53NUVRi8xXvNG99kaWygC0TpFXDXf%2FdcPfsPS8io4VyezDmoudDRc7ekLUDT%2BOCGtMbK5HlODFAHE8TT0qFXMh%2BBLJaUrJ%2BFHjmcN%2B7tjMwZbr7FUjvszEqIm71dht90KFkhvXRZeR3z4CP5F0qVH0j%2Fnjrj2Zlfg7NLwazYzfRa78HBK95JK6Di2imC69hk27N%2BYII6CMcLJmM7tvaYHp90tpCcFWH1yoILr3SvZRFA8k%2BiZtwxikEy7pU44SSVVR027nZI65tkEiR4JN00qrkUVYMNm76NCRKaUZlizpp3TQJ8ghRlFDFOojQ66oL9rikxuiKqfgImRu%2BgI1PqX7lxWg%3D%3D; type=REGISTERED; wm_ul_plus=INACTIVE|1616648469715; ONEAPP_CUSTOMER=true; member=0; athrvi=RVI~ha4d66df; _abck=hqwoigx515i3t5ob07av_1782; _pxvid=d728ca08-8c40-11eb-bf0e-0242ac120007; tb_sw_supported=false; TBV=13; TB_Latency_Tracker_100=1; TB_Navigation_Preload_01=1; TB_DC_Flap_Test=1; TB_Latency_Tracker_100=1; TB_Navigation_Preload_01=1; TB_SFOU-100=; TS012c809b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; TS01af768b=01538efd7c9969f3722db49594ea5106654dd8e60954598498e7aac183f690bb136973b988d57d8baf0ca261fceca4cbb5bd7b1295; bstc=SvIK5kMsYLTEqtq0yE3j7Y; exp-ck=44e3B14644E1Aa-Hd1vZriO1; g=0; mobileweb=0; vtc=S51vv5NXKeZVTw7VVq68Mc; xpa=44e3B|4644E|Aa-Hd|GLQ0p|vZriO; xpm=7%2B1616564245%2BbxpoIRrPw1pR5N4Jb6zHxo~dabc72e4-8920-424d-be60-842367422ceb%2B0; TS01bae75b=01538efd7c6b7b00aebf23763a1ddc5619c4e4dfe0276b30b5a16831a4d8c6ebdb5b4ed08af619d621174c92848b8b86206bb0855f; x-csrf-jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiY29va2llIiwidXVpZCI6IjlmNTNjM2QwLThjNjMtMTFlYi1hNjE3LTRmODE0OTQ3OGM1OCIsImlhdCI6MTYxNjU2NDUwNiwiZXhwIjoxNjE3NjQ0NTA2fQ.Hy75SDUpi3ck8VzNSnh8Lc_TIDOmyNd9g1HGzqFQfKU")
		// req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		req.Header.Add("Host", "www.walmart.com")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Safari/605.1.15")
		req.Header.Add("Referer", "https://www.walmart.com/grocery/")
		req.Header.Add("Accept-Language", "en-us")
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("wg-correlation-id", "8a11ded0-8c63-11eb-ad85-290b60e46528")
		req.Header.Add("x-csrf-jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiaGVhZGVyIiwidXVpZCI6ImViMzlhN2MwLThjNjItMTFlYi05ZGY5LTY1NDFjOTBlOTRhYiIsImlhdCI6MTYxNjU2NDIwNCwiZXhwIjozMjY2ODU2NTg5NX0.lBxXzxG5cOCmHyInz5s8jIn_tkbiEtn5R54JSR1gaVM")
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(body))
	return string(body)
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

// Get the login token from a gin context and verify its integrity
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
