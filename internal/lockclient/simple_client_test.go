package lockclient

import (
	"testing"
	"time"

	"github.com/SystemBuilders/LocKey/internal/cache"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

// func TestRaft(t *testing.T) {
// 	raftLS := lockservice.New(true)
// 	// raftLS, err := lockservice.NewRaftServer(
// 	// 	"test",
// 	// 	"127.0.0.1:5000",
// 	// )
// 	raftLS.RaftAddr = "127.0.0.1:5000"
// 	raftLS.Open(true, "node0")
// 	raftLS.Start()

// 	scfg := lockservice.NewSimpleConfig("http://127.0.0.1", "5001")

// 	size := 5
// 	cache := cache.NewLRUCache(size)
// 	sc := NewSimpleClient(*scfg, *cache)

// 	d := lockservice.NewSimpleDescriptor("test", "owner")

// 	time.Sleep(3 * time.Second)

// 	got := sc.Acquire(d)
// 	var want error
// 	if got != want {
// 		t.Errorf("acquire: got %q want %q", got, want)
// 	}

// 	got = sc.Release(d)

// }

func TestAcquireandRelease(t *testing.T) {
	// zerolog.New(os.Stdout).With()

	// log := zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())
	// scfg := lockservice.NewSimpleConfig("http://127.0.0.1", "1234")
	// ls := lockservice.NewSimpleLockService(log)
	scfg := lockservice.NewSimpleConfig("http://127.0.0.1", "5001")

	quit := make(chan bool, 1)
	go func() {
		raftLS := lockservice.New(true)

		raftLS.RaftAddr = "127.0.0.1:5000"
		raftLS.Open(true, "node0")
		raftLS.Start()

		for {
			select {
			case <-quit:
				return
			default:
			}
		}
	}()

	// Server takes some time to start
	time.Sleep(3 * time.Second)
	t.Run("acquire test release test", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(*scfg, *cache)

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
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(*scfg, *cache)

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
		size := 1
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(*scfg, *cache)

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
			t.Errorf("release: got %v want %v", got, want)
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
	quit <- true
	return
}
