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
	duration := 2 * time.Second // 2 second expiry
	ls := lockservice.NewSimpleLockService(log, duration)

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

	// Server takes some time to start.
	time.Sleep(100 * time.Millisecond)

	// Flow of creating a client and acquiring a lock:
	// 1. Create a cache for the client.
	// 2. Create a client and plug in the created cache.
	// 3. Connect to the said client and hold on to the session value.
	// 4. Use the session as a key for all further transactions.
	t.Run("acquire test release test", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, log, cache)

		session := sc.Connect()

		d := lockservice.NewObjectDescriptor("test")

		got := sc.Acquire(d, session)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewObjectDescriptor("test1")
		got = sc.Acquire(d, session)
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewObjectDescriptor("test")
		got = sc.Release(d, session)
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}

		d = lockservice.NewObjectDescriptor("test1")
		got = sc.Release(d, session)
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("acquire test, acquire test, release test", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, log, cache)

		session := sc.Connect()
		d := lockservice.NewObjectDescriptor("test")

		got := sc.Acquire(d, session)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		session2 := sc.Connect()
		got = sc.Acquire(d, session2)
		want = lockservice.ErrFileacquired
		if got.Error() != want.Error() {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Release(d, session)
		want = nil
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("acquire test, trying to release test as another entity should fail", func(t *testing.T) {
		size := 2
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, log, cache)

		session := sc.Connect()
		d := lockservice.NewObjectDescriptor("test")
		got := sc.Acquire(d, session)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		session2 := sc.Connect()
		got = sc.Release(d, session2)
		want = lockservice.ErrUnauthorizedAccess
		if got != want {
			t.Errorf("acquire: got %v want %v", got, want)
		}

		d = lockservice.NewObjectDescriptor("test2")
		got = sc.Acquire(d, session)
		want = nil
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewObjectDescriptor("test")
		got = sc.Release(d, session)
		want = nil
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}

		d = lockservice.NewObjectDescriptor("test2")
		got = sc.Release(d, session)
		want = nil
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("acquire test and release after session expiry", func(t *testing.T) {
		sc := NewSimpleClient(scfg, log, nil)
		session := sc.Connect()
		d := lockservice.NewObjectDescriptor("test3")

		got := sc.Acquire(d, session)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		// Wait for the session to expire
		time.Sleep(500 * time.Millisecond)
		got = sc.Release(d, session)
		want = ErrSessionNonExistent
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
	ls := lockservice.NewSimpleLockService(log, 5)

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

	sc := NewSimpleClient(scfg, log, nil)
	session := sc.Connect()
	d := lockservice.NewObjectDescriptor("test")
	for n := 0; n < b.N; n++ {
		got := sc.Acquire(d, session)
		var want error
		if got != want {
			b.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Release(d, session)
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
	duration := 2 * time.Second // 2 second expiry
	ls := lockservice.NewSimpleLockService(log, duration)

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
	sc := NewSimpleClient(scfg, log, cache)
	session := sc.Connect()
	d := lockservice.NewObjectDescriptor("test")
	for n := 0; n < b.N; n++ {
		got := sc.Acquire(d, session)
		var want error
		if got != want {
			b.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Release(d, session)
		if got != want {
			b.Errorf("release: got %q want %q", got, want)
		}
	}
}
