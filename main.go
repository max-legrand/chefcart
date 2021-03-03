/*
Package main ...
	Calls instance of webapp and runs server
*/
package main

import (
	"fmt"
	"main/webapp"
	"os/exec"
)

func main() {
	fmt.Println("Starting up...")
	out, _ := exec.Command("ls", "-a").Output()
	fmt.Println(string(out))
	webapp.LaunchServer()
	// webapp.Prototest()
}
