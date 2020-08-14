package lockclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/SystemBuilders/LocKey/internal/cache"
	"github.com/SystemBuilders/LocKey/internal/consensus"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

func TestAcquireandRelease(t *testing.T) {
	scfg := lockservice.NewSimpleConfig("http://127.0.0.1", "5001")
	scfg1 := lockservice.NewSimpleConfig("http://127.0.0.1", "7001")

	quit := make(chan bool, 1)
	go func() {
		raftLS := consensus.New(true)

		raftLS.RaftAddr = "127.0.0.1:5000"
		raftLS.Open(true, "node0")
		raftLS.Start()

		time.Sleep(3 * time.Second)
		raftLS2 := consensus.New(true)

		raftLS2.RaftAddr = "127.0.0.1:6000"
		raftLS2.Open(true, "node2")
		raftLS2.Start()

		fmt.Printf("joining")
		raftLS.Join("127.0.0.1:6000", "node2")
		time.Sleep(5 * time.Second)

		time.Sleep(3 * time.Second)

		raftLS3 := consensus.New(true)

		raftLS3.RaftAddr = "127.0.0.1:7000"
		raftLS3.Open(true, "node3")
		raftLS3.Start()

		time.Sleep(3 * time.Second)

		fmt.Printf("joining")
		raftLS.Join("127.0.0.1:7000", "node3")
		time.Sleep(5 * time.Second)

		for {
			select {
			case <-quit:
				return
			default:
			}
		}
	}()

	// Server takes some time to start
	time.Sleep(30 * time.Second)
	t.Run("acquire test release test", func(t *testing.T) {
		size := 5
		cache := cache.NewLRUCache(size)
		sc := NewSimpleClient(*scfg, *cache)
		sc1 := NewSimpleClient(*scfg1, *cache)

		d := lockservice.NewSimpleDescriptor("test", "owner")

		got := sc.Acquire(d)
		time.Sleep(5 * time.Second)
		var want error
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test1", "owner")

		got = sc.Acquire(d)
		time.Sleep(5 * time.Second)
		if got != want {
			t.Errorf("acquire: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test", "owner")

		got = sc1.Release(d)
		time.Sleep(5 * time.Second)
		if got != want {
			t.Errorf("release: got %q want %q", got, want)
		}

		d = lockservice.NewSimpleDescriptor("test1", "owner")

		got = sc1.Release(d)
		time.Sleep(5 * time.Second)
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
