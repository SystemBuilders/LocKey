package lockclient

import (
	"testing"

	"github.com/GoPlayAndFun/LocKey/internal/lockservice"
)

func TestAcquire(t *testing.T) {
	sc := SimpleClient{config: SimpleConfig{IPAddr: "http://127.0.0.1", PortAddr: "61111"}}
	d := lockservice.NewSimpleDescriptor("test")

	got := sc.Acquire(d)
	var want error

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

}

func TestRelease(t *testing.T) {
	sc := SimpleClient{config: SimpleConfig{IPAddr: "http://127.0.0.1", PortAddr: "61111"}}
	d := lockservice.NewSimpleDescriptor("test")

	got := sc.Release(d)
	var want error

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

}
