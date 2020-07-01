package lockclient

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/GoPlayAndFun/LocKey/internal/lockservice"
)

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
	Id     uint8
	config *SimpleConfig
}

// Acquire makes a HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Acquire(d lockservice.Descriptors) error {
	baseUrl := sc.config.IPAddr + ":" + sc.config.PortAddr
	endPoint := baseUrl + "/acquire"

	var jsonStr = []byte(`{"FileID":` + string(d) + "}")
	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(jsonStr))
	req.Header.set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if string(body) == "lock acquired" {
		return nil
	}

}

// Release makes a HTTP call to the lockserver and acquires the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
func (sc *SimpleClient) Release(d lockservice.Descriptors) error {
	baseUrl := sc.config.IPAddr + ":" + sc.config.PortAddr
	endPoint := baseUrl + "/release"

	var jsonStr = []byte(`{"FileID":` + string(d) + "}")
	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(jsonStr))
	req.Header.set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if string(body) == "lock released" {
		return nil
	}

}

// StartService starts the lockservice Lockey.
// This creates a new instance of the service and then starts the server.
func (sc *SimpleClient) StartService(cfg Config) error {
	panic("TODO")
}
