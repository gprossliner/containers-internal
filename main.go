package main

import (
	"fmt"
	"os"
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
}
