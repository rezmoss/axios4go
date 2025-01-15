package main

import (
	"fmt"

	"github.com/rezmoss/axios4go"
)

func main() {
	axios4go.GetAsync("https://api.github.com/users/rezmoss").
		Then(func(response *axios4go.Response) {
			// Handle successful response
			var user map[string]interface{}
			err := response.JSON(&user)
			if err != nil {
				fmt.Printf("Error parsing JSON: %v\n", err)
				return
			}
			fmt.Printf("GitHub User: %s\n", user["login"])
			fmt.Printf("Name: %s\n", user["name"])
			fmt.Printf("Public Repos: %v\n", user["public_repos"])
		}).
		Catch(func(err error) {
			// Handle error
			fmt.Printf("Error occurred: %v\n", err)
		}).
		Finally(func() {
			// This will always be executed
			fmt.Println("Request completed")
		})
}
