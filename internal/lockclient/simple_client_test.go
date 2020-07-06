package lockclient

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/GoPlayAndFun/LocKey/internal/lockservice"
	"github.com/GoPlayAndFun/LocKey/internal/node"

	"github.com/rs/zerolog"
)

func TestAcquireandRelease(t *testing.T) {
	zerolog.New(os.Stdout).With()

	log := zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())
	scfg := lockservice.NewSimpleConfig("127.0.0.1", "1234")
	ls := lockservice.NewSimpleLockService(log)

	quit := make(chan bool)
	go func() {
		node.Start(ls, *scfg)
		for {
			select {
			case <-quit:
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
				return
			default:

			}
		}

	}()

	time.Sleep(100 * time.Millisecond)
	t.Run("acquire 'test'", func(t *testing.T) {
		sc := SimpleClient{config: SimpleConfig{IPAddr: "http://127.0.0.1", PortAddr: "1234"}}
		d := lockservice.NewSimpleDescriptor("test")

		got := sc.Acquire(d)
		var want error

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}

	})

	t.Run("release 'test'", func(t *testing.T) {
		sc := SimpleClient{config: SimpleConfig{IPAddr: "http://127.0.0.1", PortAddr: "1234"}}
		d := lockservice.NewSimpleDescriptor("test")

		got := sc.Release(d)
		var want error

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}

	})
	fmt.Println("hello1")
	quit <- true
	node.StopChannel()

	fmt.Println("hello2")
	return

}
