package main

import (
	"log"
	"teriyake/go-git-it/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("cmd.Execute err: %s", err)
	}
}
