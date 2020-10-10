package lockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/SystemBuilders/LocKey/internal/lockclient/id"

	"github.com/SystemBuilders/LocKey/internal/lockclient/cache"
	"github.com/SystemBuilders/LocKey/internal/lockclient/session"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

var _ Config = (*lockservice.SimpleConfig)(nil)

// SimpleClient implements Client, the lockclient for LocKey.
type SimpleClient struct {
	config *lockservice.SimpleConfig
	cache  *cache.LRUCache
	mu     sync.Mutex
	id     id.ID

	// sessions holds the mapping of a process to a session.
	sessions map[id.ID]session.Session
	// sessionTimers maintains the timers for each session,
	sessionTimers map[id.ID]chan struct{}
	// sessionAcquisitions has a list of all the acquisitions
	// from a particular process. This has no knowledge of
	// whether the process owning the lock has an active session
	// or not, this guarantee has to be ensured by the client.
	sessionAcquisitions map[id.ID][]lockservice.Descriptors
}

// NewSimpleClient returns a new SimpleClient of the given parameters.
// This client works with or without the existance of a cache.
func NewSimpleClient(config *lockservice.SimpleConfig, cache *cache.LRUCache) *SimpleClient {
	clientID := id.Create()
	sessions := make(map[id.ID]session.Session)
	sessionTimers := make(map[id.ID]chan struct{})
	sessionAcquisitions := make(map[id.ID][]lockservice.Descriptors)
	return &SimpleClient{
		config:              config,
		cache:               cache,
		id:                  clientID,
		sessions:            sessions,
		sessionTimers:       sessionTimers,
		sessionAcquisitions: sessionAcquisitions,
	}
}

var _ Client = (*SimpleClient)(nil)

// Connect lets the user process to establish a connection with the
// client.
func (sc *SimpleClient) Connect() session.Session {
	sessionID := id.Create()
	processID := id.Create()
	session := session.NewSession(sessionID, sc.id, processID)
	sc.sessions[processID] = session
	sc.startSession(processID)
	return session
}

// Acquire allows the user process to acquire a lock.
// This returns a "session expired" error if the session expires when
// the lock is being acquired.
//
// All locks acquired during the session will be revoked if the session
// expires.
func (sc *SimpleClient) Acquire(d lockservice.Object, s session.Session) error {
	if _, ok := sc.sessions[s.ProcessID()]; !ok {
		return ErrSessionInexistent
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			sc.mu.Lock()
			select {
			case <-sc.sessionTimers[s.ProcessID()]:
				cancel()
				sc.gracefulSessionShutDown(s.ProcessID())
			default:
			}
			sc.mu.Unlock()
		}
	}()
	ld := lockservice.NewLockDescriptor(d.ID(), s.ProcessID().String())
	err := sc.acquire(ctx, ld)
	if err != nil {
		return err
	}
	// Once the lock is guaranteed to be acquired, append it to the acquisitions list.
	sc.mu.Lock()
	sc.sessionAcquisitions[s.ProcessID()] = append(sc.sessionAcquisitions[s.ProcessID()], ld)
	sc.mu.Unlock()
	return nil
}

// acquire makes an HTTP call to the lockserver and acquires the lock.
// This function makes the acquire call and doesn't care about the contention
// on the lock service.
// The errors involved may be due the HTTP, cache or the lockservice errors.
//
// This function doesn't care about sessions or ordering of the user processes and
// thus can be used for book-keeping purposes using a nil context.
func (sc *SimpleClient) acquire(ctx context.Context, d lockservice.Descriptors) (err error) {

	if ctx != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					err = SessionExpired
					return
				default:
				}
			}
		}()
	}

	// Check for existance of a cache and check
	// if the element is in the cache.
	if sc.cache != nil {
		_, err := sc.getFromCache(lockservice.ObjectDescriptor{ObjectID: d.ID()})
		// Since there can be cache errors, we have this double check.
		// We need to exit if a cache doesn't exist but proceed if the cache
		// failed in persisting this element.
		if err != nil && err != lockservice.ErrCheckAcquireFailure {
			return err
		}
	}

	endPoint := sc.config.IP() + ":" + sc.config.Port() + "/acquire"
	// Since the cache doesn't have the element, query the server.
	testData := lockservice.LockRequest{FileID: d.ID(), UserID: d.Owner()}
	requestJSON, ok := json.Marshal(testData)
	if ok != nil {
		return ok
	}

	req, ok := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestJSON))
	if ok != nil {
		return ok
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, ok := client.Do(req)
	if ok != nil {
		return ok
	}
	defer resp.Body.Close()

	body, ok := ioutil.ReadAll(resp.Body)
	if ok != nil {
		return ok
	}
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

// Release makes an HTTP call to the lockserver and releases the lock.
// The errors invloved may be due the HTTP errors or the lockservice errors.
//
// Only if there is an active session by the user process, it can release the locks
// once verified that the locks belong to the user process.
func (sc *SimpleClient) Release(d lockservice.Object, s session.Session) error {
	if _, ok := sc.sessions[s.ProcessID()]; !ok {
		return ErrSessionInexistent
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			sc.mu.Lock()
			select {
			case <-sc.sessionTimers[s.ProcessID()]:
				cancel()
				sc.gracefulSessionShutDown(s.ProcessID())
			default:
			}
			sc.mu.Unlock()
		}
	}()
	ld := lockservice.NewLockDescriptor(d.ID(), s.ProcessID().String())
	err := sc.release(ctx, ld)
	if err != nil {
		return err
	}
	// Remove the descriptor that was released.
	sc.removeFromSlice(s.ProcessID(), ld)
	return nil
}

// release makes a HTTP call to the lock service and releases the lock.
// This function makes the release call and doesn't care about the contention
// on the lock service.
// The errors involved maybe the HTTP, cache or the lockservice errors.
//
// This function doesn't care about sessions or ordering of the user processes and
// thus can be used for book-keeping purposes using a nil context.
// TODO: Cache invalidation
func (sc *SimpleClient) release(ctx context.Context, d lockservice.Descriptors) (err error) {

	if ctx != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					err = SessionExpired
					return
				default:
				}
			}
		}()
	}

	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/release"
	data := lockservice.LockRequest{FileID: d.ID(), UserID: d.Owner()}
	requestJSON, ok := json.Marshal(data)
	if ok != nil {
		return ok
	}

	req, ok := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestJSON))
	if ok != nil {
		return ok
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, ok := client.Do(req)
	if ok != nil {
		return ok
	}
	defer resp.Body.Close()

	body, ok := ioutil.ReadAll(resp.Body)
	if ok != nil {
		return ok
	}
	if resp.StatusCode != 200 {
		return lockservice.Error(strings.TrimSpace(string(body)))
	}

	if sc.cache != nil {
		err := sc.releaseFromCache(d)
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

// CheckAcquire checks for acquisition of lock and returns the owner if the lock
// is already acquired.
// The errors returned can be due to HTTP errors or marshalling errors.
// A "file is not acquired" error is returned if so and no error and an owner is
// returned if the object is acquired.
func (sc *SimpleClient) CheckAcquire(d lockservice.ObjectDescriptor) (string, error) {
	if sc.cache != nil {
		return sc.getFromCache(d)
	}

	endPoint := sc.config.IPAddr + ":" + sc.config.PortAddr + "/checkAcquire"
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
func (sc *SimpleClient) getFromCache(d lockservice.ObjectDescriptor) (string, error) {
	if sc.cache != nil {
		owner, err := sc.cache.GetElement(cache.NewSimpleKey(d.ObjectID, ""))
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

// startSession starts the session by initiating the timer for this user process.
// This is a non blocking function which runs on a different goroutine. It sends
// a signal through the "sessionTimers" map for the respective "processID" when
// the session timer ends.
//
// The function starts with creating a new channel, assigning it to the respective
// object in the map and then ends by closing the channel created.
func (sc *SimpleClient) startSession(processID id.ID) {
	go func(id.ID) {
		timerChan := make(chan struct{}, 1)
		sc.mu.Lock()
		sc.sessionTimers[processID] = timerChan
		sc.mu.Unlock()
		// Sessions last for 200ms.
		time.Sleep(200 * time.Millisecond)
		sc.mu.Lock()
		sc.sessionTimers[processID] <- struct{}{}
		sc.mu.Unlock()
		close(sc.sessionTimers[processID])
	}(processID)
}

// gracefulSessionShutdown releases all the locks in the lockservice once the
// session has ended.
func (sc *SimpleClient) gracefulSessionShutDown(processID id.ID) {
	for i := range sc.sessionAcquisitions[processID] {
		sc.release(nil, sc.sessionAcquisitions[processID][i])
	}
}

func (sc *SimpleClient) removeFromSlice(processID id.ID, d lockservice.Descriptors) {
	sc.mu.Lock()
	for i := range sc.sessionAcquisitions[processID] {
		if sc.sessionAcquisitions[processID][i] == d {
			sc.sessionAcquisitions[processID] = append(sc.sessionAcquisitions[processID][:i], sc.sessionAcquisitions[processID][i+1:]...)
		}
	}
	sc.mu.Unlock()
}
