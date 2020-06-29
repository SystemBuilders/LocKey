package lockclient

import "github.com/GoPlayAndFun/LocKey/internal/lockservice"

var _ Config = (*SimpleConfig)(nil)

// SimpleConfig implements Config.
type SimpleConfig struct {
	IPAddr   string
	PortAddr string
}

// IP returns the IP from SimpleConfig.
func (scfg *SimpleConfig) IP() string {
	return scfg.IPAddr
}

// Port returns the port from SimpleConfig.
func (scfg *SimpleConfig) Port() string {
	return scfg.PortAddr
}

var _ Client = (*SimpleClient)(nil)

// SimpleClient implements Client, the lockclient for Lockey.
type SimpleClient struct {
}

// Acquire makes a HTTP call to the lockserver and acquires the lock.
func (sc *SimpleClient) Acquire(d lockservice.Descriptors) error {
	panic("TODO")
}

// Release makes a HTTP call to the lockserver and acquires the lock.
func (sc *SimpleClient) Release(d lockservice.Descriptors) error {
	panic("TODO")
}

// StartService starts the lockservice Lockey.
func (sc *SimpleClient) StartService(cfg Config) error {
	panic("TODO")
}
