/*
Package models ...
	Utilizes (G)ORM to convert database entries into usable Go Structs
*/
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

// UserInfo ...
type UserInfo struct {
	gorm.Model
	City         string
	State        string
	Restirctions pq.StringArray `gorm:"type:character varying[]"`
}

// User ...
type User struct {
	gorm.Model
	Email    string
	Password string
	UserInfo UserInfo `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// type Usertest struct {
// 	gorm.Model
// 	Name      string
// 	CompanyID int
// 	Company   Company
// }

// type Company struct {
// 	ID   int
// 	Name string
// }

// DB ...
var DB *gorm.DB

// ConnectDB ...
func ConnectDB() {
	godotenv.Load(".env")
	// database, err := gorm.Open("sqlite3", "test.db")
	// postgres://omntajdhdpgnrw:80fc2161b99b8ca8a6a045f5a30bc5fc771a81c748d181d7f8d5db76ef84c0b2@ec2-18-207-95-219.compute-1.amazonaws.com:5432/dc96vunmur80be
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

	database.AutoMigrate(&User{}, &UserInfo{})
	// database.AutoMigrate(&Usertest{})
	// database.AutoMigrate(&Company{})
	DB = database

	// comptest := Company{ID: 100, Name: "Mycompnay"}
	// DB.Create(&Usertest{Name: "test", Company: comptest})

}
