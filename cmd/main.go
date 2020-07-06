package main

import (
	"os"

	"github.com/GoPlayAndFun/LocKey/internal/lockservice"
	"github.com/GoPlayAndFun/LocKey/internal/node"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.New(os.Stdout).With()

	log := zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())
	ls := lockservice.NewSimpleLockService(log)

	scfg := lockservice.NewSimpleConfig("127.0.0.1", "61111")
	node.Start(ls, *scfg)
}
