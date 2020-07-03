package lockclient

import (
	"os"
	"testing"

	"github.com/GoPlayAndFun/LocKey/internal/lockservice"
	"github.com/GoPlayAndFun/LocKey/internal/node"
	"github.com/rs/zerolog"
)

func TestAcquire(t *testing.T) {

	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				zerolog.New(os.Stdout).With()

				log := zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())
				scfg := lockservice.SimpleConfig{
					IPAddr:   "127.0.0.1",
					PortAddr: "1234",
				}
				ls := lockservice.NewSimpleLockService(log)
				node.Start(ls, scfg)
			}
		}
	}()

	go func() {

		for {
			select {
			case <-quit:
				return
			default:
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
			}
		}

	}()
	quit <- true

}
