package lockclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/SystemBuilders/LocKey/internal/cache"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

var _ Config = (*lockservice.SimpleConfig)(nil)

// SimpleClient implements Client, the lockclient for LocKey.
type SimpleClient struct {
	config *lockservice.SimpleConfig
	cache  *cache.LRUCache
	mu     sync.Mutex
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

// Acquire allows the user process to acquire a lock.
func (sc *SimpleClient) Acquire(d lockservice.Descriptors) error {
	return sc.acquire(d)
}

// acquire makes an HTTP call to the lockserver and acquires the lock.
// This function makes the acquire call and doesn't care about the contention
// on the lock service.
// The errors involved may be due the HTTP errors or the lockservice errors.
//
// Currently acquire doesn't order the user processes that request for the lock.
func (sc *SimpleClient) acquire(d lockservice.Descriptors) error {
	// Check for existance of a cache and check
	// if the element is in the cache.
	if sc.cache != nil {
		_, err := sc.getFromCache(lockservice.ObjectDescriptor{ObjectID: d.ID()})
		// Since there can be network errors, we have this double check.
		if err != nil && err != lockservice.ErrCheckacquireFailure {
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

// Checkacquire checks for acquisition of lock and returns the owner if the lock
// is already acquired.
// The errors returned can be due to HTTP errors or marshalling errors.
// A "file is not acquired" error is returned if so and no error and an owner is
// returned if the object is acquired.
func (sc *SimpleClient) Checkacquire(d lockservice.ObjectDescriptor) (string, error) {
	if sc.cache != nil {
		owner, err := sc.getFromCache(d)
		if err != nil {
			return "", err
		}
		return owner, nil
	}

	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/checkacquire"
	data := lockservice.LockCheckRequest{FileID: d.ObjectID}
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

	var ownerData lockservice.CheckacquireRes
	err = json.Unmarshal(body, &ownerData)
	if err != nil {
		return "", err
	}

	return ownerData.Owner, nil
}

// getFromCache checks the lock status on the descriptor in the cache.
// This function returns an error if the cache doesn't exist or the
// file is NOT acquired.
func (sc *SimpleClient) getFromCache(d lockservice.ObjectDescriptor) (string, error) {
	if sc.cache != nil {
		owner, err := sc.cache.GetElement(cache.NewSimpleKey(d.ObjectID, ""))
		if err != nil {
			return "", lockservice.ErrCheckacquireFailure
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
