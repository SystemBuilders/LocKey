package lockclient

import (
	"os"
	"testing"
	"time"

	"github.com/SystemBuilders/LocKey/internal/cache"
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
	t.Run("acquire 'test' release 'test'", func(t *testing.T) {
		sc := SimpleClient{config: *scfg, cache: *cache.NewLRUCache(5)}
		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test", "owner")

		got = sc.Release(d)

		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("acquire 'test', acquire 'test', release 'test'", func(t *testing.T) {
		sc := SimpleClient{config: *scfg, cache: *cache.NewLRUCache(5)}
		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Acquire(d)
		want = ErrElementAlreadyLocked
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test", "owner")

		got = sc.Release(d)
		want = nil
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})
	quit <- true
	return
}
