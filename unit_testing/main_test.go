/*
Package main_test ...
	tests the funcitonality of the other program components
*/
package main_test

//search recipes

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	webapp "main/webapp"
	"main/webapp/controller"
	"main/webapp/models"
	Tokens "main/webapp/protobuf"
	"main/webapp/service"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
)

var tokenString string

func TestSignup(t *testing.T) {
	models.ConnectDB()
	email := "unit_test_user@unittest.com"
	password := "unit_test_user"
	data := []byte(password)
	hash := md5.Sum(data)
	newpass := hex.EncodeToString(hash[:])
	users := []models.User{}
	models.DB.Where("email = ?", email).Find(&users)
	var user models.User
	if len(users) == 0 {
		user = models.User{Email: email, Password: newpass}
		models.DB.Create(&user)
		userinfo := models.UserInfo{QuantityThreshold: -1, Diets: []string{}, Intolerances: []string{}, City: "", State: "", ID: int(user.ID)}
		models.DB.Create(&userinfo)
	} else {
		t.Errorf("User already exists")
	}
	checkUser := models.User{}
	models.DB.Where("email = ? and password = ?", email, newpass).First(&checkUser)
	if checkUser.ID == user.ID && checkUser.Email == user.Email && checkUser.Password == user.Password {
		fmt.Println("Passed User Signup")
		models.DB.Delete(&user)
	} else {
		models.DB.Delete(&user)
		t.Errorf("Failed user signup")
	}
}

func TestLogin(t *testing.T) {
	jwtService := service.JWTAuthService()
	loginController := controller.LoginHandler(jwtService)
	username := Tokens.Token{Token: "demo_user_unit_test@unittest.com"}
	password := Tokens.Token{Token: "demo_user_unit_test"}
	token := webapp.GetLoginTokenUnwrapped(loginController, &username, &password)
	if token == "" {
		t.Errorf("Error logging in user")
	} else {
		fmt.Println("Successfully logged in user")
		// fmt.Println(token)
		tokenString = strings.ReplaceAll(token, " ", "+")
	}
}

func TestAuthUser(t *testing.T) {
	// demo_user_unit_test@unittest.com
	// demo_user_unit_test
	input := Tokens.Token{Token: tokenString}
	result, err := webapp.AuthUserUnwrapped(&input)
	if err != nil {
		t.Errorf("Login user encountered an error")
	} else if result.Token == "" {
		t.Errorf("Invalid credentials supplided")
	} else {
		fmt.Println("Successfully authenticated in user")
		// fmt.Println(result.Token)
	}
}

func TestEditUser(t *testing.T) {
	input := Tokens.Token{Token: tokenString}
	result, err := webapp.AuthUserUnwrapped(&input)
	if err == nil && result.Token != "" {
		// Update user state to CA
		user := models.User{}
		userinfo := models.UserInfo{}
		models.DB.Where("email = ?", result.Token).First(&user)
		models.DB.Where("ID = ?", user.ID).First(&userinfo)
		userinfo.State = "CA"
		models.DB.Save(&user)
		models.DB.Save(&userinfo)
		// Verify change made
		newuser := models.User{}
		newuserinfo := models.UserInfo{}
		models.DB.Where("email = ?", result.Token).First(&newuser)
		models.DB.Where("ID = ?", user.ID).First(&newuserinfo)
		if newuserinfo.State != "CA" {
			t.Errorf("Error editing user")
		} else {
			fmt.Println("User successfully edited")
		}
		// Revert change
		userinfo.State = "NJ"
		models.DB.Save(&user)
		models.DB.Save(&userinfo)
	} else {
		t.Errorf("Error authenticating user")
	}
}

func TestGetPantry(t *testing.T) {
	results, err := webapp.GetPantryUnwrapped(&Tokens.Token{Token: tokenString})
	if err != nil && err.Error() == "Invalid user" {
		t.Errorf("Invalid user detected")
	} else {
		fmt.Println("Pantry recieved")
		fmt.Println(results)
	}
}

func TestAddPantry(t *testing.T) {
	results, err := webapp.AuthUserUnwrapped(&Tokens.Token{Token: tokenString})
	if err != nil && err.Error() == "Invalid user" {
		t.Errorf("Invalid user detected")
	} else {
		user := models.User{}
		models.DB.Where("email = ?", results.Token).First(&user)
		date := "1/1/1999"
		name := "cheese"

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

		image := "Default"
		if len(gjson.Get(jsonString, "annotations").Array()) == 0 {
			t.Errorf("Invalid food item")
			return
		}
		if image == "Default" {
			image = gjson.Get(jsonString, "annotations.0.image").String()
		}
		quantity := "1"
		volume := "1 fl.oz"
		weight := "1 g"
		item := models.Ingredient{
			Name:       name,
			UID:        user.ID,
			Quantity:   quantity,
			Volume:     volume,
			Weight:     weight,
			ImageLink:  image,
			Expiration: date,
		}
		models.DB.Create(&item)

		newPantryItem := models.Ingredient{}
		models.DB.Where("name = ? and UID = ?", "cheese", user.ID).Find(&newPantryItem)
		if newPantryItem.Name == item.Name && newPantryItem.UID == item.UID && newPantryItem.Quantity == item.Quantity && newPantryItem.Volume == item.Volume && newPantryItem.Weight == item.Weight && newPantryItem.Expiration == item.Expiration && newPantryItem.ImageLink == item.ImageLink {
			fmt.Println("Item created successfully")
			models.DB.Delete(&item)
		} else {
			models.DB.Delete(&item)
			t.Errorf("Item not created successfully")
		}
	}
}

func TestVerifyFood(t *testing.T) {
	name := "car"

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
	jsonString := string(body)

	if len(gjson.Get(jsonString, "annotations").Array()) == 0 {
		fmt.Println("Success - invalid food item detected")
		return
	} else {
		t.Errorf("Failed food detection")
	}
}

func TestEditItem(t *testing.T) {
	results, err := webapp.AuthUserUnwrapped(&Tokens.Token{Token: tokenString})
	if err == nil && results.Token != "" {
		// Seed item
		user := models.User{}
		models.DB.Where("email = ?", results.Token).First(&user)
		item := models.Ingredient{
			Name:       "cheese",
			Quantity:   "1",
			Weight:     "1 g",
			Volume:     "1 fl.oz",
			Expiration: "1/1/1999",
			ImageLink:  "",
			UID:        user.ID,
		}
		models.DB.Create(&item)
		item.Name = "updated_cheese"
		models.DB.Save(&item)
		newitem := models.Ingredient{}
		models.DB.Where("ID = ? and name = ? and UID = ?", item.ID, "updated_cheese", user.ID).Find(&newitem)
		if newitem.Name == "updated_cheese" && newitem.UID == item.UID && newitem.Quantity == item.Quantity && newitem.Volume == item.Volume && newitem.Weight == item.Weight && newitem.Expiration == item.Expiration && newitem.ImageLink == item.ImageLink {
			fmt.Println("Item updated successfully")
			models.DB.Delete(&item)
		} else {
			t.Errorf("Item failed to update")
			models.DB.Delete(&item)
		}
	} else {
		t.Errorf("Invalid user")
	}
}

func TestDelete(t *testing.T) {
	results, err := webapp.AuthUserUnwrapped(&Tokens.Token{Token: tokenString})
	if err == nil && results.Token != "" {
		// Seed item
		user := models.User{}
		models.DB.Where("email = ?", results.Token).First(&user)
		item := models.Ingredient{
			Name:       "cheese",
			Quantity:   "1",
			Weight:     "1 g",
			Volume:     "1 fl.oz",
			Expiration: "1/1/1999",
			ImageLink:  "",
			UID:        user.ID,
		}
		models.DB.Create(&item)
		res := models.DB.Delete(&item)
		if res.RowsAffected == 1 {
			fmt.Println("Successfully deleted item")
		} else {
			t.Errorf("Failed to delete item")
		}
	} else {
		t.Errorf("Invalid user")
	}
}

func TestFindRecipes(t *testing.T) {
	godotenv.Load(".env")
	ingredients := "cheese"
	offset := 0
	url := "https://api.spoonacular.com/recipes/complexSearch?&includeIngredients=" + ingredients + "&number=10&offset=" + strconv.Itoa(offset) + "&apiKey=" + os.Getenv("APIKEY")
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
	if totalResults > 0 {
		fmt.Println("Found recipes")
	} else {
		t.Errorf("Could not find recipes")
	}
}

func TestGroceries(t *testing.T) {
	results, err := webapp.AuthUserUnwrapped(&Tokens.Token{Token: tokenString})
	user := models.User{}
	models.DB.Where("email = ?", results.Token).First(&user)
	input := Tokens.Token{Token: tokenString}
	result, err := webapp.GetGroceries(&input)
	if err != nil {
		t.Errorf("Encountered unexpected error")
	}
	if len(result.Pantry) != 0 {
		t.Errorf("Got pantry item, expected empty list")
	}
	item := models.Grocery{
		UID:  user.ID,
		Name: "testfood",
	}
	models.DB.Create(&item)
	result, err = webapp.GetGroceries(&input)
	if err != nil {
		t.Errorf("Encountered unexpected error")
	}
	if len(result.Pantry) == 0 {
		t.Errorf("Got pantry item, expected empty list")
	}
	models.DB.Delete(&item)
	fmt.Println("Succesfully retrieved groceries")
}

func TestGroceriesResults(t *testing.T) {
	results, _ := webapp.AuthUserUnwrapped(&Tokens.Token{Token: tokenString})
	user := models.User{}
	models.DB.Where("email = ?", results.Token).First(&user)
	item := models.Grocery{
		UID:  user.ID,
		Name: "cheese",
	}
	models.DB.Create(&item)
	in := Tokens.SearchQuery{
		Token: tokenString,
		ID:    fmt.Sprint(item.ID),
	}
	resultsStore, _ := webapp.GetSearchResults(&in)
	if resultsStore.Address == "" {
		t.Errorf("Failed to retrieve results")
	}
	models.DB.Delete(&item)
	fmt.Println("Succesfully retrieved groceries from search")
	fmt.Println(fmt.Sprint(len(resultsStore.Results)) + " Results found")
}
