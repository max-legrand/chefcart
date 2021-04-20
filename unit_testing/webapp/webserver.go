/*
Package webapp ...
	Runs webserver and displays content
*/
package webapp

import (
	"errors"
	"fmt"
	"io/ioutil"
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
	"github.com/tidwall/gjson"
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

// GetGroceries - return grocerys for user
func GetGroceries(in *Tokens.Token) (*Tokens.Pantry, error) {
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

// GetSearchResults - returns results for grocery inquiry
func GetSearchResults(in *Tokens.SearchQuery) (*Tokens.Store, error) {
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
