package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}

	must(cmd.Run())
}

func subdir(parent string, name string) string {
	dir := path.Join(parent, name)
	must(os.MkdirAll(dir, os.ModePerm))
	return dir
}

func child() {

	// initialize a overlay mount for our container-fs
	lowerdir := "/tmp/alpineroot"
	tmpdir, err := os.MkdirTemp("", "containers-internal-")
	must(err)
	fmt.Printf("Created tmp dir %s\n", tmpdir)
	defer os.RemoveAll(tmpdir)

	upperdir := subdir(tmpdir, "upper")
	workdir := subdir(tmpdir, "workdir")
	rootdir := subdir(tmpdir, "root")

	// mount overlayfs
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerdir, upperdir, workdir)
	must(syscall.Mount("overlay", rootdir, "overlay", 0, options))

	// change the root for the child process
	must(syscall.Chroot(rootdir))
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
