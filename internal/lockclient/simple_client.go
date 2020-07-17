package lockclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

var _ Config = (*SimpleConfig)(nil)

// SimpleConfig implements Config.
type SimpleConfig struct {
	IPAddr   string
	PortAddr string
}

// SimpleClient implements Client, the lockclient for LocKey.
type SimpleClient struct {
	config SimpleConfig
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

// Acquire makes a HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Acquire(d lockservice.Descriptors) error {
	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/acquire"

	testData := lockservice.LockRequest{FileID: d.ID()}
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
	return nil
}

// Release makes a HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Release(d lockservice.Descriptors) error {
	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/release"

	testData := lockservice.LockRequest{FileID: d.ID()}
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

	return nil
}

// StartService starts the lockservice LocKey.
// This creates a new instance of the service and then starts the server.
func (sc *SimpleClient) StartService(cfg Config) error {
	panic("TODO")
}
