package lockservice

import (
	"sync"

	"github.com/rs/zerolog"
)

// SafeLockMap is the lockserver's data structure
type SafeLockMap struct {
	LockMap map[string]string
	Mutex   sync.Mutex
}

// SimpleConfig implements Config.
type SimpleConfig struct {
	IPAddr   string
	PortAddr string
}

// LockRequest is an instance of a request for a lock.
type LockRequest struct {
	FileID string `json:"FileID"`
	UserID string `json:"UserID"`
}

// LockCheckRequest is an instance of a lock check request.
type LockCheckRequest struct {
	FileID string `json:"FileID"`
}

// CheckAcquireRes is the response of a CheckAcquire.
type CheckAcquireRes struct {
	Owner string `json:"owner"`
}

// IP returns the IP from the SimpleConfig.
func (scfg *SimpleConfig) IP() string {
	return scfg.IPAddr
}

// Port returns the port from SimpleConfig.
func (scfg *SimpleConfig) Port() string {
	return scfg.PortAddr
}

var _ LockService = (*SimpleLockService)(nil)

// SimpleLockService is a lock service that implements LockService.
// It uses a golang map to maintain the locks of the descriptors.
// It can acquire and release locks and has an in-built logger.
type SimpleLockService struct {
	log     zerolog.Logger
	lockMap *SafeLockMap
}

var _ Descriptors = (*SimpleDescriptor)(nil)

// SimpleDescriptor implements the Descriptors interface.
// Many descriptors can be added to this struct and the ID
// can be a combination of all those descriptors.
type SimpleDescriptor struct {
	FileID string
	UserID string
}

// ID represents the distinguishable ID of the descriptor.
func (sd *SimpleDescriptor) ID() string {
	return sd.FileID
}

// Owner represents the distinguishable ID of the entity that
// holds the lock for FileID.
func (sd *SimpleDescriptor) Owner() string {
	return sd.UserID
}

// NewSimpleConfig returns an instance of the SimpleConfig.
func NewSimpleConfig(IPAddr, PortAddr string) *SimpleConfig {
	return &SimpleConfig{
		IPAddr:   IPAddr,
		PortAddr: PortAddr,
	}
}

// NewSimpleDescriptor returns an instance of the SimpleDescriptor.
func NewSimpleDescriptor(FileID, UserID string) *SimpleDescriptor {
	return &SimpleDescriptor{
		FileID: FileID,
		UserID: UserID,
	}
}

// NewSimpleLockService creates and returns a new lock service ready to use.
func NewSimpleLockService(log zerolog.Logger) *SimpleLockService {
	safeLockMap := &SafeLockMap{
		LockMap: make(map[string]string),
	}
	return &SimpleLockService{
		log:     log,
		lockMap: safeLockMap,
	}
}

// Acquire function lets a client acquire a lock on an object.
func (ls *SimpleLockService) Acquire(sd Descriptors) error {
	ls.lockMap.Mutex.Lock()
	if _, ok := ls.lockMap.LockMap[sd.ID()]; ok {
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("can't acquire, already been acquired")
		return ErrFileAcquired
	}
	ls.lockMap.LockMap[sd.ID()] = sd.Owner()
	ls.lockMap.Mutex.Unlock()
	ls.
		log.
		Debug().
		Str("descriptor", sd.ID()).
		Msg("locked")
	return nil
}

// Release lets a client to release a lock on an object.
func (ls *SimpleLockService) Release(sd Descriptors) error {
	ls.lockMap.Mutex.Lock()
	// Only the entity that posseses the lock for this object
	// is allowed to release the lock
	if ls.lockMap.LockMap[sd.ID()] == sd.Owner() {
		delete(ls.lockMap.LockMap, sd.ID())
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("released")
		ls.lockMap.Mutex.Unlock()
		return nil
	} else if _, ok := ls.lockMap.LockMap[sd.ID()]; !ok {
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("can't release, hasn't been acquired")
		ls.lockMap.Mutex.Unlock()
		return ErrCantReleaseFile

	} else {
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("can't release, unauthorized access")
		ls.lockMap.Mutex.Unlock()
		return ErrUnauthorizedAccess

	}

}

// CheckAcquired returns true if the file is acquired.
// It also returns the owner of the file.
func (ls *SimpleLockService) CheckAcquired(sd Descriptors) (string, bool) {
	ls.lockMap.Mutex.Lock()
	id := sd.ID()
	if owner, ok := ls.lockMap.LockMap[id]; ok {
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", id).
			Msg("checkAcquire success")
		return owner, true
	}
	ls.
		log.
		Debug().
		Str("descriptor", id).
		Msg("check Acquire failure")
	ls.lockMap.Mutex.Unlock()
	return "", false
}

// CheckReleased returns true if the file is released
func (ls *SimpleLockService) CheckReleased(sd Descriptors) bool {
	ls.lockMap.Mutex.Lock()
	id := sd.ID()
	if _, ok := ls.lockMap.LockMap[id]; ok {
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", id).
			Msg("checkRelease failure")
		return false
	}
	ls.lockMap.Mutex.Unlock()
	ls.
		log.
		Debug().
		Str("descriptor", id).
		Msg("checkRelease success")
	return true
}
