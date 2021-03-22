package main

import(
	"os"
	"testing"
)


func Test_CommandHelp(t *testing.T) {
	args := os.Args[0:1]
	args = append(args, "-h") 
	run(args)
}