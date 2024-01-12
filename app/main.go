package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var perm fs.FileMode = 0755

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	image := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:]

	sandbox, err := os.MkdirTemp("", "sandbox-")
	must(err)

	Download(image, sandbox)

	// copy the command from outside the sandbox, preserving the permissions
	/* copyFile(command, sandbox+command) */

	err = syscall.Chroot(sandbox)
	must(err)
	os.Chdir("/")

	createDevNull()

	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// pid namespace separation
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

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
