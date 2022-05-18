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
	switch os.Args[1] {
	case "run":
		run()
	default:
		panic("invalid args")
	}
}

func run() {
	fmt.Printf("Running %v UID=%d PID=%d\n", os.Args[1:], os.Geteuid(), os.Getpid())

	// create a exec Command and setup pipes
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}
	must(cmd.Run())
}
