package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	command := os.Args[3]
	args := os.Args[4:]

	sandbox, err := os.MkdirTemp("", "sandbox")
	must(err)

	// copy the command from outside the sandbox, preserving the permissions
	copyFile(command, sandbox+command)

	syscall.Chroot(sandbox)
	os.Chdir("/")

	createDevNull()

	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		statusCode := exitErr.ExitCode()
		os.Exit(statusCode)
	}
	must(err)
}

func copyFile(srcFile, destFile string) {
	src, err := os.Open(srcFile)
	must(err)
	defer src.Close()

	err = os.MkdirAll(filepath.Dir(destFile), 0755)
	must(err)

	dest, err := os.Create(destFile)
	must(err)
	defer dest.Close()

	_, err = io.Copy(dest, src)
  must(err)

  stat, err := src.Stat()
	must(err)

  err = dest.Chmod(stat.Mode())
  must(err)
}

func createDevNull() {
	// to avoid problem running in chroot
	os.Mkdir("/dev", 0755)
	devNull, _ := os.Create("/dev/null")
	devNull.Close()
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
