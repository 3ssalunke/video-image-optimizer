package main

import (
	"log"

	"github.com/3ssalunke/vio/vio/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
