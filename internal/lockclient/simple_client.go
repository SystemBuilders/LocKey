package lockclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/SystemBuilders/LocKey/internal/cache"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

var _ Config = (*lockservice.SimpleConfig)(nil)

// SimpleClient implements Client, the lockclient for LocKey.
type SimpleClient struct {
	config *lockservice.SimpleConfig
	cache  *cache.LRUCache
}

// NewSimpleClient returns a new SimpleKey of the given value.
// This client works with or without the existance of a cache.
func NewSimpleClient(config *lockservice.SimpleConfig, cache *cache.LRUCache) *SimpleClient {
	return &SimpleClient{
		config: config,
		cache:  cache,
	}
}

var _ Client = (*SimpleClient)(nil)

// Acquire makes an HTTP call to the lockserver and acquires the lock.
// The errors involved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Acquire(d lockservice.Descriptors) error {
	endPoint := sc.config.IP() + ":" + sc.config.Port() + "/acquire"
	err := sc.getFromCache(d)
	if err != nil {
		return err
	}

	testData := lockservice.LockRequest{FileID: d.ID(), UserID: d.Owner()}
	requestJSON, err := json.Marshal(testData)

	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return (err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New(strings.TrimSpace(string(body)))
	}

	err = sc.releaseFromCache(d)
	if err != nil {
		return err
	}
	return nil
}

// Release makes an HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Release(d lockservice.Descriptors) error {
	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/release"
	testData := lockservice.LockRequest{FileID: d.ID(), UserID: d.Owner()}
	requestJSON, err := json.Marshal(testData)
	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return (err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return lockservice.Error(strings.TrimSpace(string(body)))
	}

	return nil
}

// StartService starts the lockservice LocKey.
// This creates a new instance of the service and then starts the server.
func (sc *SimpleClient) StartService(cfg Config) error {
	panic("TODO")
}

// Watch can be used to watch the given lock.
// This works with or without the existance of a cache
// for the client.
//
// On calling Watch, the current state of the lock is
// returned. If the lock is not acquired, the function returns.
// If the lock is acquired by a different process, the
// details of the acquirer is sent back and the function
// doesn't return unless explicitly told to. Only on changes
// in the ownership of the lock, details are sent back to
// the function caller.
// The function also returns if THIS owner eventually pounced
// and got access to the lock.
func (sc *SimpleClient) Watch(d lockservice.Descriptors, quit chan struct{}) (chan Lock, error) {
	stateChan := make(chan Lock, 1)
	if sc.cache != nil {
		err := sc.getFromCache(d)
		if err != nil {
			return nil, err
		}
		// Send the initial state of the lock and then
		// keep sending state changes until stopped
		// explicitly.
		stateChan <- Lock{d.Owner(), Acquire}
		go func() {
			for {
				select {
				case <-quit:
					return
				default:
					<-time.After(100 * time.Millisecond)
					err := sc.getFromCache(d)
					if err != nil {
						return
					}
				}
			}
		}()
		return stateChan, nil
	}

	return nil, nil
}

// Pounce can be used to pounce on a waiting lock.
func (sc *SimpleClient) Pounce(lockservice.Descriptors) error {
	panic("TODO")
}

// Pouncers can be used to check the existing pouncers on a descriptor.
func (sc *SimpleClient) Pouncers(lockservice.Descriptors) []string {
	panic("TODO")
}

// getFromCache checks the lock status on the descriptor in the cache.
func (sc *SimpleClient) getFromCache(d lockservice.Descriptors) error {
	if sc.cache != nil {
		err := sc.cache.GetElement(cache.NewSimpleKey(d.ID()))
		if err == nil {
			return lockservice.ErrFileAcquired
		}
		return nil
	}
	return cache.ErrCacheDoesntExist
}

func (sc *SimpleClient) releaseFromCache(d lockservice.Descriptors) error {
	if sc.cache != nil {
		err := sc.cache.RemoveElement(cache.NewSimpleKey(d.ID()))
		if err != nil {
			return err
		}
		return nil
	}
	return cache.ErrCacheDoesntExist
}
