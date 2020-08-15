package lockclient

import (
	"os"
	"testing"
	"time"

	"github.com/SystemBuilders/LocKey/internal/cache"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
	"github.com/SystemBuilders/LocKey/internal/node"
	"github.com/stretchr/testify/assert"

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

		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test1", "owner")

		got = sc.acquire(d)
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test", "owner")

		got = sc.Release(d)

		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test1", "owner")

		got = sc.Release(d)

		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("acquire test, acquire test, release test", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.acquire(d)
		want = lockservice.ErrFileacquired
		if got.Error() != want.Error() {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test", "owner")

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

		d := lockservice.NewSimpleDescriptor("test", "owner1")
		got := sc.acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test", "owner2")
		got = sc.Release(d)
		want = lockservice.ErrUnauthorizedAccess
		if got != want {
			t.Errorf("acquire: got %v want %v", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test2", "owner1")
		got = sc.acquire(d)
		want = nil
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test", "owner1")
		got = sc.Release(d)
		want = nil
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}
	})

	t.Run("lock watching without cache", func(t *testing.T) {
		sc := NewSimpleClient(scfg, nil)

		assert := assert.New(t)
		d := lockservice.NewSimpleDescriptor("test", "owner1")
		// acquire the lock
		err := sc.acquire(d)
		assert.Nil(err)

		// start watching the lock.
		quit := make(chan struct{}, 1)
		stateChan, err := sc.Watch(d, quit)
		assert.Nil(err)

		states := []Lock{}
		go func() {
			for {
				state := <-stateChan
				states = append(states, state)
			}
		}()

		err = sc.Release(d)
		assert.Nil(err)

		d1 := lockservice.NewSimpleDescriptor("test", "owner2")
		err = sc.acquire(d1)
		assert.Nil(err)

		err = sc.Release(d1)
		assert.Nil(err)

		d2 := lockservice.NewSimpleDescriptor("test", "owner3")
		err = sc.acquire(d2)
		assert.Nil(err)

		err = sc.Release(d2)
		assert.Nil(err)

		quit <- struct{}{}

		// wait to stop watching gracefully.
		<-time.After(10 * time.Millisecond)
	})

	t.Run("lock watching with cache", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		assert := assert.New(t)
		d := lockservice.NewSimpleDescriptor("test", "owner1")
		// acquire the lock
		err := sc.acquire(d)
		assert.Nil(err)

		// start watching the lock.
		quit := make(chan struct{}, 1)
		stateChan, err := sc.Watch(d, quit)
		assert.Nil(err)

		states := []Lock{}
		go func() {
			for {
				state := <-stateChan
				states = append(states, state)
			}
		}()

		err = sc.Release(d)
		assert.Nil(err)

		d1 := lockservice.NewSimpleDescriptor("test", "owner2")
		err = sc.acquire(d1)
		assert.Nil(err)

		err = sc.Release(d1)
		assert.Nil(err)

		d2 := lockservice.NewSimpleDescriptor("test", "owner3")
		err = sc.acquire(d2)
		assert.Nil(err)

		err = sc.Release(d2)
		assert.Nil(err)

		quit <- struct{}{}

		// wait to stop watching gracefully.
		<-time.After(10 * time.Millisecond)
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
	d := lockservice.NewSimpleDescriptor("test", "owner")
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
	d := lockservice.NewSimpleDescriptor("test", "owner")
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
