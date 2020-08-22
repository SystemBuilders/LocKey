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

		d := lockservice.NewLockDescriptor("test", "owner")

		got := sc.acquire(d)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewLockDescriptor("test1", "owner")

		got = sc.acquire(d)
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
		got := sc.acquire(d)
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
		got = sc.acquire(d)
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

	t.Run("lock watching without cache", func(t *testing.T) {
		sc := NewSimpleClient(scfg, nil)

		assert := assert.New(t)
		d := lockservice.NewLockDescriptor("test", "owner1")
		// acquire the lock
		err := sc.acquire(d)
		assert.Nil(err)

		// start watching the lock.
		quit := make(chan struct{}, 1)
		stateChan, err := sc.Watch(lockservice.ObjectDescriptor{ObjectID: d.ID()}, quit)
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

		d1 := lockservice.NewLockDescriptor("test", "owner2")
		err = sc.acquire(d1)
		assert.Nil(err)

		err = sc.Release(d1)
		assert.Nil(err)

		d2 := lockservice.NewLockDescriptor("test", "owner3")
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
		d := lockservice.NewLockDescriptor("test", "owner1")
		// acquire the lock
		err := sc.acquire(d)
		assert.Nil(err)

		// start watching the lock.
		quit := make(chan struct{}, 1)
		stateChan, err := sc.Watch(lockservice.ObjectDescriptor{ObjectID: d.ID()}, quit)
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

		d1 := lockservice.NewLockDescriptor("test", "owner2")
		err = sc.acquire(d1)
		assert.Nil(err)

		err = sc.Release(d1)
		assert.Nil(err)

		d2 := lockservice.NewLockDescriptor("test", "owner3")
		err = sc.acquire(d2)
		assert.Nil(err)

		err = sc.Release(d2)
		assert.Nil(err)

		quit <- struct{}{}

		// wait to stop watching gracefully.
		<-time.After(10 * time.Millisecond)
	})

	// This test first makes a process acquire a lock.
	// Later, pounces using 3 different owners on the same object,
	// and then release one by one and observe the behaviour and
	// assert the expected behaviour.
	// Time waits are added in order to maintain the sequence of
	// pounces and have a deterministic test.
	t.Run("pounce test without quitting in between, always wait for object", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		assert := assert.New(t)
		d := lockservice.NewLockDescriptor("test", "owner")
		err := sc.Acquire(d)
		assert.NoError(err)

		objD := lockservice.NewObjectDescriptor("test")

		go func() {
			err := sc.Pounce(*objD, "owner1", nil, true)
			assert.NoError(err)
		}()

		go func() {
			<-time.After(100 * time.Millisecond)
			err := sc.Pounce(*objD, "owner2", nil, true)
			assert.NoError(err)
		}()

		go func() {
			<-time.After(500 * time.Millisecond)
			err := sc.Pounce(*objD, "owner3", nil, true)
			assert.NoError(err)
		}()

		pouncersBeforeFirstRelease := []string{"owner1", "owner2", "owner3"}
		pouncersAfterFirstRelease := []string{"owner2", "owner3"}
		pouncersAfterSecondRelease := []string{"owner3"}
		pouncersAfterThirdRelease := []string{}

		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersBeforeFirstRelease)

		err = sc.Release(d)
		assert.NoError(err)
		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersAfterFirstRelease)

		<-time.After(1 * time.Second)
		owner, err := sc.Checkacquire(*objD)
		if owner == "owner1" {
			err = sc.Release(&lockservice.LockDescriptor{FileID: objD.ObjectID, UserID: "owner1"})
			assert.NoError(err)
		}

		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersAfterSecondRelease)

		<-time.After(1 * time.Second)
		owner, err = sc.Checkacquire(*objD)
		if owner == "owner2" {
			err = sc.Release(&lockservice.LockDescriptor{FileID: objD.ObjectID, UserID: "owner2"})
			assert.NoError(err)
		}

		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersAfterThirdRelease)
	})

	// This test is very similar to the one above but owner2 doensn't wait for its pounce.
	t.Run("pounce test without quitting in between, without waiting for object", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		assert := assert.New(t)
		d := lockservice.NewLockDescriptor("test1", "owner")
		err := sc.Acquire(d)
		assert.NoError(err)

		objD := lockservice.NewObjectDescriptor("test1")

		go func() {
			err := sc.Pounce(*objD, "owner1", nil, true)
			assert.NoError(err)
		}()

		go func() {
			<-time.After(100 * time.Millisecond)
			err := sc.Pounce(*objD, "owner2", nil, false)
			if err != ErrorObjectAlreadyPouncedOn {
				assert.NoError(err)
			}
		}()

		go func() {
			<-time.After(500 * time.Millisecond)
			err := sc.Pounce(*objD, "owner3", nil, true)
			assert.NoError(err)
		}()

		pouncersBeforeFirstRelease := []string{"owner1", "owner3"}
		pouncersAfterFirstRelease := []string{"owner3"}
		pouncersAfterSecondRelease := []string{}

		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersBeforeFirstRelease)

		err = sc.Release(d)
		assert.NoError(err)
		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersAfterFirstRelease)

		<-time.After(1 * time.Second)
		owner, err := sc.Checkacquire(*objD)
		if owner == "owner1" {
			err = sc.Release(&lockservice.LockDescriptor{FileID: objD.ObjectID, UserID: "owner1"})
			assert.NoError(err)
		}

		<-time.After(2 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersAfterSecondRelease)
	})

	// This test is very similar to the one above but a quit signal is sent to owner2's pounce.
	t.Run("pounce test with quitting in between, without waiting for object", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(scfg, cache)

		assert := assert.New(t)
		d := lockservice.NewLockDescriptor("testx", "owner")

		quitChan := make(chan struct{}, 1)
		err := sc.Acquire(d)
		assert.NoError(err)

		objD := lockservice.NewObjectDescriptor("testx")

		go func() {
			err := sc.Pounce(*objD, "owner1", nil, true)
			assert.NoError(err)
		}()

		go func() {
			<-time.After(100 * time.Millisecond)
			err := sc.Pounce(*objD, "owner2", quitChan, true)
			assert.NoError(err)
		}()

		go func() {
			<-time.After(500 * time.Millisecond)
			err := sc.Pounce(*objD, "owner3", nil, true)
			assert.NoError(err)
		}()

		pouncersBeforeFirstRelease := []string{"owner1", "owner2", "owner3"}
		pouncersAfterFirstRelease := []string{"owner3"}
		pouncersAfterSecondRelease := []string{}

		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersBeforeFirstRelease)

		quitChan <- struct{}{}
		err = sc.Release(d)
		assert.NoError(err)
		<-time.After(1 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersAfterFirstRelease)

		<-time.After(1 * time.Second)
		owner, err := sc.Checkacquire(*objD)
		if owner == "owner1" {
			err = sc.Release(&lockservice.LockDescriptor{FileID: objD.ObjectID, UserID: "owner1"})
			assert.NoError(err)
		}

		<-time.After(2 * time.Second)
		assert.Equal(sc.Pouncers(*objD), pouncersAfterSecondRelease)
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
