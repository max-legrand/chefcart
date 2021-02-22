/*
Package models ...
	Utilizes (G)ORM to convert database entries into usable Go Structs
*/
package models

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"

	// sqlite
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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
	database, err := gorm.Open("sqlite3", "test.db")
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
