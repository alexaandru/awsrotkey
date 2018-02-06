package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alexaandru/awsrotkey/internal/rotate"
)

type config struct {
	dryMode bool
	profile string
}

var cfg = &config{profile: os.Getenv("AWS_PROFILE")}

func main() {
	flag.Parse()

	if err := rotate.Key(cfg.profile, cfg.dryMode); err != nil {
		fmt.Println(err)
		os.Exit(42)
	}
}

func init() {
	flag.BoolVar(&cfg.dryMode, "dry", cfg.dryMode, "Dry mode: preview only")
	flag.StringVar(&cfg.profile, "profile", cfg.profile, "The profile to use")
}
