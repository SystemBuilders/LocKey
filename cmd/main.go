package main

import (
	"os"

	"github.com/SystemBuilders/LocKey/internal/lockservice"
	"github.com/SystemBuilders/LocKey/internal/lockservice/node"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.New(os.Stdout).With()

	log := zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())
	ls := lockservice.NewSimpleLockService(log)

	scfg := lockservice.NewSimpleConfig("127.0.0.1", "1234")
	node.Start(ls, *scfg)
}
