package lockclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/SystemBuilders/LocKey/internal/cache"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
	"github.com/rs/zerolog/log"
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
	// Check for existance of a cache and check
	// if the element is in the cache.
	if sc.cache != nil {
		_, err := sc.getFromCache(d)
		// Since there can be network errors, we have this double check.
		if err != nil && err != lockservice.ErrCheckAcquireFailure {
			return err
		}
	}

	endPoint := sc.config.IP() + ":" + sc.config.Port() + "/acquire"
	// Since the cache doesn't have the element, query the server.
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
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New(strings.TrimSpace(string(body)))
	}

	if sc.cache != nil {
		err := sc.addToCache(d)
		if err != nil {
			return err
		}
	}
	return nil
}

// Release makes an HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Release(d lockservice.Descriptors) error {
	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/release"
	data := lockservice.LockRequest{FileID: d.ID(), UserID: d.Owner()}
	requestJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return lockservice.Error(strings.TrimSpace(string(body)))
	}

	if sc.cache != nil {
		err = sc.releaseFromCache(d)
		if err != nil {
			return err
		}
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
//
// The function also returns if THIS owner eventually pounced
// and got access to the lock.
//
// The stateChan is assumed to be cleared as soon as data is
// sent through it.
func (sc *SimpleClient) Watch(d lockservice.Descriptors, quit chan struct{}) (chan Lock, error) {
	stateChan := make(chan Lock, 1)
	// releaseNotification is true if the last notification wasn't a release.
	releaseNotification := false
	owner, err := sc.CheckAcquire(d)
	if err != nil {
		if err != lockservice.ErrCheckAcquireFailure {
			return nil, err
		}
		// This means that the file is released
		if releaseNotification {
			releaseNotification = false
			stateChan <- Lock{"", Release}
			log.Debug().
				Str("process", "lock watching").
				Str("lock", d.ID()).
				Msg("lock is in released state")
		}
	}
	// Send the initial state of the lock and then
	// keep sending state changes until stopped
	// explicitly.
	if owner != "" {
		releaseNotification = true
		stateChan <- Lock{owner, Acquire}
		log.Debug().
			Str("process", "lock watching").
			Str("lock", d.ID()).
			Str("owner", owner).
			Msg("lock is in acquired state")
	}

	go func() {
		for {
			select {
			case <-quit:
				log.Debug().Msg("stopped watching")
				return
			default:
				newOwner, err := sc.CheckAcquire(d)
				if err != nil {
					if err != lockservice.ErrCheckAcquireFailure {
						return
					}
					if releaseNotification {
						releaseNotification = false
						stateChan <- Lock{"", Release}
						log.Debug().
							Str("process", "lock watching").
							Str("lock", d.ID()).
							Msg("lock is in released state")
					}
				} else {
					// notify about the state only if there's a change.
					if newOwner != owner {
						releaseNotification = true
						owner = newOwner
						stateChan <- Lock{owner, Acquire}
						log.Debug().
							Str("process", "lock watching").
							Str("lock", d.ID()).
							Str("owner", owner).
							Msg("lock is in acquired state")
					}
				}
			}
		}
	}()
	return stateChan, nil
}

// Pounce can be used to pounce on a waiting lock.
func (sc *SimpleClient) Pounce(lockservice.Descriptors) error {
	panic("TODO")
}

// Pouncers can be used to check the existing pouncers on a descriptor.
func (sc *SimpleClient) Pouncers(lockservice.Descriptors) []string {
	panic("TODO")
}

// CheckAcquire checks for acquisition of lock and returns the owner if the lock
// is already acquired.
func (sc *SimpleClient) CheckAcquire(d lockservice.Descriptors) (string, error) {
	if sc.cache != nil {
		owner, err := sc.getFromCache(d)
		if err != nil {
			return "", err
		}
		return owner, nil
	}

	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/checkAcquire"
	data := lockservice.LockCheckRequest{FileID: d.ID()}
	requestJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", lockservice.Error(strings.TrimSpace(string(body)))
	}

	var ownerData lockservice.CheckAcquireRes
	err = json.Unmarshal(body, &ownerData)
	if err != nil {
		return "", err
	}

	return ownerData.Owner, nil
}

// getFromCache checks the lock status on the descriptor in the cache.
// This function returns an error if the cache doesn't exist or the
// file is NOT acquired.
func (sc *SimpleClient) getFromCache(d lockservice.Descriptors) (string, error) {
	if sc.cache != nil {
		owner, err := sc.cache.GetElement(cache.NewSimpleKey(d.ID(), d.Owner()))
		if err != nil {
			return "", lockservice.ErrCheckAcquireFailure
		}
		return owner, nil
	}
	return "", cache.ErrCacheDoesntExist
}

func (sc *SimpleClient) addToCache(d lockservice.Descriptors) error {
	if sc.cache != nil {
		err := sc.cache.PutElement(cache.NewSimpleKey(d.ID(), d.Owner()))
		if err != nil {
			return err
		}
		return nil
	}
	return cache.ErrCacheDoesntExist
}

func (sc *SimpleClient) releaseFromCache(d lockservice.Descriptors) error {
	if sc.cache != nil {
		err := sc.cache.RemoveElement(cache.NewSimpleKey(d.ID(), d.Owner()))
		if err != nil {
			return err
		}
		return nil
	}
	return cache.ErrCacheDoesntExist
}
