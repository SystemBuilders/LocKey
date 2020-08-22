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
	"github.com/rs/zerolog/log"
)

var _ Config = (*lockservice.SimpleConfig)(nil)

// SimpleClient implements Client, the lockclient for LocKey.
type SimpleClient struct {
	config   *lockservice.SimpleConfig
	cache    *cache.LRUCache
	pouncers map[lockservice.ObjectDescriptor][]string
	mu       sync.Mutex
}

// NewSimpleClient returns a new SimpleKey of the given value.
// This client works with or without the existance of a cache.
func NewSimpleClient(config *lockservice.SimpleConfig, cache *cache.LRUCache) *SimpleClient {
	p := make(map[lockservice.ObjectDescriptor][]string)
	return &SimpleClient{
		config:   config,
		cache:    cache,
		pouncers: p,
	}
}

var _ Client = (*SimpleClient)(nil)

// Acquire internally calls Pounce in order to follow the FIFO order.
// Acquire assumes that the process will wait for the lock until it's released.
func (sc *SimpleClient) Acquire(d lockservice.Descriptors) error {
	return sc.Pounce(lockservice.ObjectDescriptor{ObjectID: d.ID()}, d.Owner(), nil, true)
}

// acquire makes an HTTP call to the lockserver and acquires the lock.
// The errors involved may be due the HTTP errors or the lockservice errors.
//
// acquire can ONLY be called by Pouncer when it's sure that this was the
// process next in line to acquire the lock.
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

// Watch can be used to watch the given lock descriptor.
// This works with or without the existance of a cache
// for the client.
//
// On calling Watch, the current state of the lock is
// returned.
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
//
// Usage:
//		quitChan := make(chan struct{},1)
//		stateChan, err := client.Watch(descriptor,quitChan)
//		go func() {
//			for {
//				state := <-stateChan
//				// process the state data.
//			}
//		}
//
func (sc *SimpleClient) Watch(d lockservice.ObjectDescriptor, quit chan struct{}) (chan Lock, error) {
	stateChan := make(chan Lock, 1)
	// releaseNotification is true if the last notification wasn't a release.
	releaseNotification := false
	owner, err := sc.Checkacquire(d)
	if err != nil {
		if err != lockservice.ErrCheckacquireFailure {
			return nil, err
		}
		// This means that the file is released
		if releaseNotification {
			releaseNotification = false
			stateChan <- Lock{"", Release}
			log.Debug().
				Str("process", "lock watching").
				Str("lock", d.ObjectID).
				Msg("lock is in released state")
		}
	}
	// Send the initial state of the lock and then
	// keep sending state changes until stopped
	// explicitly.
	if owner != "" {
		releaseNotification = true
		stateChan <- Lock{owner, acquire}
		log.Debug().
			Str("process", "lock watching").
			Str("lock", d.ObjectID).
			Str("owner", owner).
			Msg("lock is in acquired state")
	}

	go func() {
		for {
			select {
			case <-quit:
				close(stateChan)
				log.Debug().Msg("stopped watching")
				return
			default:
				newOwner, err := sc.Checkacquire(d)
				if err != nil {
					if err != lockservice.ErrCheckacquireFailure {
						return
					}
					if releaseNotification {
						releaseNotification = false
						stateChan <- Lock{"", Release}
						log.Debug().
							Str("process", "lock watching").
							Str("lock", d.ObjectID).
							Msg("lock is in released state")
					}
				} else {
					// notify about the state only if there's a change.
					if newOwner != owner {
						releaseNotification = true
						owner = newOwner
						stateChan <- Lock{owner, acquire}
						log.Debug().
							Str("process", "lock watching").
							Str("lock", d.ObjectID).
							Str("owner", owner).
							Msg("lock is in acquired state")
					}
				}
			}
		}
	}()
	return stateChan, nil
}

// Pounce can be used to acquire a lock that has already been acquired.
// It allows the process to wait for the lock in a queue and obtain it
// in FCFS order.
//
// One lock can have many pouncers. The "pouncer" can choose to
// wait on the object until it gets the lock or return if there was
// a preceding "pouncer" for the lock.
//
// The "pouncer" can stop "pouncing" at any time by signalling through
// the channel passed as the argument.
//
// The boolean argument dictates the function to pounce or not
// when there is an existng "pouncer", "true" for pounce even with "pouncers".
//
// The pounce returns on achieving its goal of acquiring the lock or when
// there are no more pouncers to be served.
//
// Pounce usage:
// 		go func() {
// 			err := sc.Pounce(objectDesc,owner,quitChan,instantPounce)
//			// errcheck
// 		}
// Pounce must be used inside a goroutine as it's a blocking process.
// Handling of ErrObjectHasBeenPouncedOn is necesssary.
func (sc *SimpleClient) Pounce(d lockservice.ObjectDescriptor, owner string, quit chan struct{}, instant bool) error {

	log.Debug().
		Str("object", d.ObjectID).
		Str("pouncer", owner).
		Msg("beginning pounce")

	// Reject the pounce if the client doesn't want to pounce on
	// pre-pounced objects.
	if !instant && (len(sc.Pouncers(d)) > 0) {
		log.Debug().Msg("stopped pounce activity, object already under pounce")
		return ErrorObjectAlreadyPouncedOn
	}

	// If the lock is already acquired, the pouncee is added to the waiting list of
	// pouncers, else, the process gets access to the lock immidiately.
	_, err := sc.Checkacquire(d)
	if err == nil {
		sc.mu.Lock()
		sc.pouncers[d] = append(sc.pouncers[d], owner)
		sc.mu.Unlock()
	} else {
		return sc.acquire(&lockservice.LockDescriptor{FileID: d.ObjectID, UserID: owner})
	}

	// Here we start waiting for the quit signal in order to
	// end on command or wait on the lock state to grant access
	// to the pouncer.
	// The lock is continously watched and whenever the lock is
	// released the first pouncer obtains the lock.
	q := make(chan struct{})
	stateChan, err := sc.Watch(d, q)
	if err != nil {
		return err
	}
	for {
		select {
		case <-quit:
			// This owner will be added to the pouncer's queue,
			// it must be removed
			sc.removeOwner(d, owner)
			log.Debug().Msg("stop signal received, stopped pouncing process")
		default:
			state := <-stateChan
			// When there's a release operation that occurred, a
			// new client process can get access to the lock.
			// Always the first element in the slice is granted access
			// to the lock on the object.
			if state.LockState == Release {
				var op string
				if len(sc.Pouncers(d)) > 0 {
					op = sc.Pouncers(d)[0]
				}

				// Errors arising here aren't propagated because this
				// process doesn't care about it.
				curOwner, err := sc.Checkacquire(d)
				if curOwner == "" && err != nil {
					desc := lockservice.NewLockDescriptor(d.ObjectID, op)
					err = sc.acquire(desc)
					if err == nil {
						sc.mu.Lock()
						// Once the task of acquiring is complete, remove
						// the first element from the slice as it was granted the lock.
						sc.pouncers[d] = append(sc.pouncers[d][:0], sc.pouncers[d][1:]...)
						sc.mu.Unlock()

						log.Debug().
							Str("lock granted through pounce to", op).
							Msg("pounce success")

						// Exit condition is the pouncer getting access to the lock.
						if op == owner || len(sc.Pouncers(d)) == 0 {
							log.Debug().
								Str("owner", owner).
								Msg("stopping pouncing process, pouncer availed lock")
							q <- struct{}{}
							close(q)
							return nil
						}
					}
				}

			}
		}
	}
}

// Pouncers returns the active "pouncers" on a descriptor.
// This is safe to read concurrently.
func (sc *SimpleClient) Pouncers(d lockservice.ObjectDescriptor) []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.pouncers[d]
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

func (sc *SimpleClient) removeOwner(d lockservice.ObjectDescriptor, owner string) {
	index := -1
	sc.mu.Lock()
	for i := 0; i < len(sc.pouncers[d]); i++ {
		if sc.pouncers[d][i] == owner {
			index = i
			break
		}
	}
	if index != -1 {
		sc.pouncers[d] = append(sc.pouncers[d][:index], sc.pouncers[d][index+1:]...)
	}
	sc.mu.Unlock()
}
