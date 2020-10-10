package lockclient

import (
	"os"
	"testing"
	"time"

	"github.com/SystemBuilders/LocKey/internal/lockclient/cache"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
	"github.com/SystemBuilders/LocKey/internal/lockservice/node"

	"github.com/rs/zerolog"
)

func TestLockService(t *testing.T) {
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
	t.Run("acquire test release test", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		d := lockservice.NewLockDescriptor("test", "owner")

		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewLockDescriptor("test1", "owner")

		got = sc.Acquire(d)
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewLockDescriptor("test", "owner")

		got = sc.Release(d)

		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}

		d = lockservice.NewLockDescriptor("test1", "owner")

		got = sc.Release(d)

		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("acquire test, acquire test, release test", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		d := lockservice.NewLockDescriptor("test", "owner")

		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Acquire(d)
		want = lockservice.ErrFileacquired
		if got.Error() != want.Error() {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewLockDescriptor("test", "owner")

		got = sc.Release(d)
		want = nil
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("acquire test, trying to release test as another entity should fail", func(t *testing.T) {
		size := 2
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		d := lockservice.NewLockDescriptor("test", "owner1")
		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewLockDescriptor("test", "owner2")
		got = sc.Release(d)
		want = lockservice.ErrUnauthorizedAccess
		if got != want {
			t.Errorf("acquire: got %v want %v", got, want)
		}

		d = lockservice.NewLockDescriptor("test2", "owner1")
		got = sc.Acquire(d)
		want = nil
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewLockDescriptor("test", "owner1")
		got = sc.Release(d)
		want = nil
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	quit <- true
	return
}

// BenchmarkLocKeyWithoutCache stats:     2130	  28828088 ns/op	   15952 B/op	   190 allocs/op
func BenchmarkLocKeyWithoutCache(b *testing.B) {
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
	time.Sleep(100 * time.Millisecond)

	sc := NewSimpleClient(scfg, nil)
	d := lockservice.NewLockDescriptor("test", "owner")
	for n := 0; n < b.N; n++ {
		got := sc.acquire(d)
		var want error
		if got != want {
			b.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Release(d)
		if got != want {
			b.Errorf("release: got %q want %q", got, want)
		}
	}
}

// BenchmarkLocKeyWithCache stats: 3669	  28702266 ns/op 16048 B/op   194 allocs/op
func BenchmarkLocKeyWithCache(b *testing.B) {
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
	time.Sleep(100 * time.Millisecond)

	size := 5
	cache := cache.NewLRUCache(size)
	sc := NewSimpleClient(scfg, cache)
	d := lockservice.NewLockDescriptor("test", "owner")
	for n := 0; n < b.N; n++ {
		got := sc.acquire(d)
		var want error
		if got != want {
			b.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Release(d)
		if got != want {
			b.Errorf("release: got %q want %q", got, want)
		}
	}
}
