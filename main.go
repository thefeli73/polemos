package main

import (
	"fmt"
	"os"

	"github.com/thefeli73/polemos/mtdaws"
	"github.com/thefeli73/polemos/state"
)

func main() {
	fmt.Println("Starting Polemos")

	config, err := state.LoadConf("config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	mtdaws.IndexInstances(config)
}
