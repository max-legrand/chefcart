/*
Package main ...
	Calls instance of webapp and runs server
*/
package main

import (
	"fmt"
	"main/webapp"
)

func main() {
	fmt.Println("Starting up...")
	webapp.LaunchServer()
}
