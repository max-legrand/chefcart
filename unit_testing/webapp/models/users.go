/*
Package models ...
	Utilizes (G)ORM to convert database entries into usable Go Structs
*/
// written by: Elysia Heah
// tested by: Jonathan Wong
// debugged by: Indrasish Moitra
package models

import (
	"net/url"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/lib/pq"

	// sqlite
	// _ "github.com/jinzhu/gorm/dialects/psql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// UserInfo - struct for maintaining additional user information
type UserInfo struct {
	ID                int `gorm:"ForeignKey:ID"`
	City              string
	State             string
	Diets             pq.StringArray `gorm:"type:character varying[]"`
	Intolerances      pq.StringArray `gorm:"type:character varying[]"`
	QuantityThreshold float64
}

// User - struct to represent user object
type User struct {
	gorm.Model
	Email    string
	Password string
}

// Grocery ...
type Grocery struct {
	gorm.Model
	UID       uint `gorm:"ForeignKey:ID"`
	Name      string
	ImageLink string
}

// Ingredient ...
type Ingredient struct {
	gorm.Model
	UID        uint `gorm:"ForeignKey:ID"`
	Name       string
	Quantity   string
	Weight     string
	Volume     string
	Expiration string
	ImageLink  string
}

// DB ...
var DB *gorm.DB

// ConnectDB - connect to the database and update the global database variable
func ConnectDB() {
	godotenv.Load(".env")
	// database, err := gorm.Open("sqlite3", "test.db")
	DBURL := os.Getenv("DATABASE_URL")
	userEndIndex := strings.Index(DBURL[11:], ":") + 11
	passEndIndex := strings.Index(DBURL, "@")
	hostEndIndex := strings.Index(DBURL[passEndIndex:], ":") + passEndIndex
	username := DBURL[11:userEndIndex]
	password := DBURL[userEndIndex+1 : passEndIndex]
	host := DBURL[passEndIndex+1 : hostEndIndex]
	dbname := DBURL[hostEndIndex+6:]
	dsn := url.URL{
		User:   url.UserPassword(username, password),
		Scheme: "postgres",
		Host:   host,
		Path:   dbname,
	}
	database, err := gorm.Open("postgres", dsn.String())
	if err != nil {
		panic("Failed to connect to DB")
	}

	database.AutoMigrate(&User{}, &UserInfo{}, &Ingredient{}, &Grocery{})
	DB = database
}
