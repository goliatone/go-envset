package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/vaughan0/go-ini"
)

func main() {

	env, err := ini.LoadFile(".envset")
	if err != nil {
		fmt.Println("Error parsing envset")
		os.Exit(1)
	}

	found := false
	environment := os.Args[1]

	for name, section := range env {
		if name == environment {
			found = true
			for key, value := range section {
				os.Setenv(key, value)
			}
		}
	}
	if found == false {
		fmt.Println("Error, environment %q not found.", environment)
		os.Exit(1)
	}

	command := os.Args[3]
	args := os.Args[4:]

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
