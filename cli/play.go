package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bjorand/nombda/engine"
)

var (
	hookFile   string
	secretFile string
)

func main() {
	flag.StringVar(&hookFile, "f", "", "hook file")
	flag.StringVar(&secretFile, "s", "", "secret file")
	flag.Parse()

	if hookFile == "" {
		log.Fatal("No file specified with -f")
	}

	hookEngine := engine.NewHookEngine("")

	secrets := make(map[string]string)
	var err error
	if secretFile != "" {
		secrets, err = engine.ReadSecretFile(secretFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	for k, v := range engine.ReadSecretFromEnv() {
		secrets[k] = v
	}
	hookEngine.Secrets = secrets
	h, err := hookEngine.ReadHookFromFile(hookFile)
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
