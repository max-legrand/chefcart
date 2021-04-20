/*
Package main ...
	Calls instance of webapp and runs server
*/
// written by: Jonathan Wong
// tested by: Mark Stanik
// debugged by: Allen Chang
package main

import (
	"fmt"
	"main/webapp"
)

func main() {
	fmt.Println("Starting up...")
	webapp.LaunchServer()
}
