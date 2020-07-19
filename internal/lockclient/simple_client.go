package lockclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SystemBuilders/LocKey/internal/cache"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

var _ Config = (*lockservice.SimpleConfig)(nil)

// SimpleClient implements Client, the lockclient for LocKey.
type SimpleClient struct {
	config lockservice.SimpleConfig
	cache  cache.LRUCache
}

var _ Client = (*SimpleClient)(nil)

// Acquire makes a HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Acquire(d lockservice.Descriptors) error {
	endPoint := sc.config.IP() + ":" + sc.config.Port() + "/acquire"

	isInCache := sc.cache.GetElement(cache.NewSimpleKey(d.ID()))

	if isInCache == nil {
		fmt.Printf("%q is already locked\n", d.ID())
		return ErrElementAlreadyLocked
	}

	testData := lockservice.LockRequest{FileID: d.ID(), UserID: d.Owner()}
	requestJson, err := json.Marshal(testData)

	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestJson))
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
		return errors.New(string(body))
	}
	err = sc.cache.PutElement(cache.NewSimpleKey(d.ID()))
	if err != nil {
		return err
	}

	return nil
}

// Release makes a HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Release(d lockservice.Descriptors) error {
	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/release"

	testData := lockservice.LockRequest{FileID: d.ID(), UserID: d.Owner()}
	requestJson, err := json.Marshal(testData)
	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestJson))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return (err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return errors.New(string(body))
	}
	err = sc.cache.RemoveElement(cache.NewSimpleKey(d.ID()))
	if err != nil {
		return err
	}

	return nil
}

// StartService starts the lockservice LocKey.
// This creates a new instance of the service and then starts the server.
func (sc *SimpleClient) StartService(cfg Config) error {
	panic("TODO")
}
