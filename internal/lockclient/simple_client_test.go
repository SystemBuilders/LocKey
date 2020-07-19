package lockclient

import (
	"os"
	"testing"
	"time"

	"github.com/SystemBuilders/LocKey/internal/lockservice"
	"github.com/SystemBuilders/LocKey/internal/node"

	"github.com/rs/zerolog"
)

func TestAcquireandRelease(t *testing.T) {
	zerolog.New(os.Stdout).With()

	log := zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())
	scfg := lockservice.NewSimpleConfig("http://127.0.0.1", "1234")
	ls := lockservice.NewSimpleLockService(log)

	quit := make(chan bool, 1)
	go func() {
		node.Start(ls, *scfg)
		for {
			select {
			case <-quit:
				return
			default:
			}
		}
	}()

	// Server takes some time to start
	time.Sleep(100 * time.Millisecond)
	t.Run("acquire 'test'", func(t *testing.T) {
		sc := SimpleClient{config: *scfg}
		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
	t.Run("release 'test'", func(t *testing.T) {
		sc := SimpleClient{config: *scfg}
		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.Release(d)
		var want error

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
	quit <- true
	return
}
