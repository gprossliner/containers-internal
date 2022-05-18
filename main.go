package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// run with `go run main.go <command>`
func main() {
	fmt.Printf("Running %v UID=%d PID=%d\n", os.Args[1:], os.Geteuid(), os.Getpid())

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("invalid args")
	}
}

func run() {

	// create the child command
	args := []string{"child"}
	args = append(args, os.Args[2:]...)

	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}

	must(cmd.Run())
}

func child() {

	// change the root for the child process
	must(syscall.Chroot("/tmp/alpineroot"))
	os.Chdir("/")

	// mount proc
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	// create a exec Command and setup pipes
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}
