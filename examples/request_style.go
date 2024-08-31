package main

import (
	"fmt"
	"log"

	"github.com/rezmoss/axios4go"
)

func main() {
	resp, err := axios4go.Request("GET", "https://api.github.com/users/rezmoss")
	if err != nil {
		log.Fatalf("Error fetching user: %v", err)
	}

	var user map[string]interface{}
	err = resp.JSON(&user)
	if err != nil {
		log.Fatalf("Error parsing user JSON: %v", err)
	}

	fmt.Printf("GitHub User: %s\n", user["login"])
	fmt.Printf("Name: %s\n", user["name"])
	fmt.Printf("Public Repos: %v\n", user["public_repos"])

}
