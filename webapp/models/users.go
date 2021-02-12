/*
Package models ...
	Utilizes (G)ORM to convert database entries into usable Go Structs
*/
package models

import (
	"github.com/jinzhu/gorm"
	// sqlite
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// User ...
type User struct {
	Email    string
	Password string
}

// DB ...
var DB *gorm.DB

// ConnectDB ...
func ConnectDB() {
	database, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("Failed to connect to DB")
	}

	database.AutoMigrate(&User{})
	DB = database

}
