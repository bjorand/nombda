package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bjorand/nombda/engine"
)

var ()

func main() {
	if len(os.Args) == 1 {
		log.Fatal("No file specified")
	}
	hookFile := os.Args[1]
	h, err := engine.ReadHookFromFile(hookFile)
	if err != nil {
		panic(err)
	}
	r, err := engine.NewRun(h)
	if err != nil {
		panic(err)
	}
	h.AsyncRun(r)
	fmt.Println(r.Log())
	os.Exit(r.ExitCode)
}
