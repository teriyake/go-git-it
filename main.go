package main

import (
	"fmt"
	"log"
	"os/exec"
	"teriyake/go-git-it/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("cmd.Execute err: %s", err)
	}
}
