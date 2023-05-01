package main

import (
	simulation "github.com/timlind/alien-invasion/simulation"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		panic(any("not enough arguments, world file and number of aliens required."))
	}

	numAliens, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		panic(any("Invalid number of aliens " + os.Args[2]))
	}

	world, err := simulation.ParseWorld(os.Args[1], numAliens)
	if err != nil {
		log.Fatal(err)
	}
	world.StartWar()
	log.Print(world.String())
}
