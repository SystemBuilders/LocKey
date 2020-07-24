package lockclient

import (
	"fmt"
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
		t.SkipNow()
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test1", "owner")

		got = sc.Acquire(d)
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
		t.SkipNow()
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.Acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		got = sc.Acquire(d)
		want = lockservice.ErrFileAcquired
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
		t.SkipNow()
		size := 1
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		d := lockservice.NewSimpleDescriptor("test", "owner1")
		got := sc.Acquire(d)
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
		got = sc.Acquire(d)
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
		// t.SkipNow()
		sc := NewSimpleClient(scfg, nil)

		assert := assert.New(t)
		d := lockservice.NewSimpleDescriptor("test", "owner1")
		// acquire the lock
		err := sc.Acquire(d)
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
		err = sc.Acquire(d1)
		assert.Nil(err)

		err = sc.Release(d1)
		assert.Nil(err)

		d2 := lockservice.NewSimpleDescriptor("test", "owner3")
		err = sc.Acquire(d2)
		assert.Nil(err)

		err = sc.Release(d2)
		assert.Nil(err)

		quit <- struct{}{}

		// wait to stop watching gracefully.
		<-time.After(10 * time.Millisecond)

		fmt.Println(states)
	})
	quit <- true
	return
}
