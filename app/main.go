package main

import (
	"log"
	"os"
	"os/exec"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	command := os.Args[3]
	args := os.Args[4:]

	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		statusCode := exitErr.ExitCode()
		os.Exit(statusCode)
	}
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
