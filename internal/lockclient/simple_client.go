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
)

var _ Config = (*lockservice.SimpleConfig)(nil)

// SimpleClient implements Client, the lockclient for LocKey.
type SimpleClient struct {
	config lockservice.SimpleConfig
	cache  cache.LRUCache
}

// NewSimpleClient returns a new SimpleKey of the given value.
func NewSimpleClient(config lockservice.SimpleConfig, cache cache.LRUCache) *SimpleClient {
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

	isInCache := sc.cache.GetElement(cache.NewSimpleKey(d.ID()))

	if isInCache == nil {
		return lockservice.ErrFileAcquired
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
	err = sc.cache.PutElement(cache.NewSimpleKey(d.ID()))
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
	sc.cache.RemoveElement(cache.NewSimpleKey(d.ID()))
	return nil
}

// StartService starts the lockservice LocKey.
// This creates a new instance of the service and then starts the server.
func (sc *SimpleClient) StartService(cfg Config) error {
	panic("TODO")
}

// Watch can be used to watch the given lock.
func (sc *SimpleClient) Watch(lockservice.Descriptors) error {
	panic("TODO")
}

// Pounce can be used to pounce on a waiting lock.
func (sc *SimpleClient) Pounce(lockservice.Descriptors) error {
	panic("TODO")
}

// Pouncers can be used to check the existing pouncers on a descriptor.
func (sc *SimpleClient) Pouncers(lockservice.Descriptors) []string {
	panic("TODO")
}
