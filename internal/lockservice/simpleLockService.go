package lockservice

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// SafeLockMap is the lockserver's data structure
type SafeLockMap struct {
	LockMap       map[string]*LockMapEntry
	LeaseDuration time.Duration
	Mutex         sync.Mutex
}

// LockMapEntry defines the structure for objects placed
// in the LockMap. It consists of the owner of the lock
// that is acquired and the timestamp at which the
// acquisition took place.
type LockMapEntry struct {
	owner     string
	timestamp time.Time
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

// CheckAcquireRes is the response of a Checkacquire.
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

var _ Descriptors = (*LockDescriptor)(nil)

// ObjectDescriptor describes the object that is subjected to
// lock operations.
type ObjectDescriptor struct {
	ObjectID string
}

// LockDescriptor implements the Descriptors interface.
// Many descriptors can be added to this struct and the ID
// can be a combination of all those descriptors.
type LockDescriptor struct {
	FileID string
	UserID string
}

// ID represents the distinguishable ID of the descriptor.
func (sd *LockDescriptor) ID() string {
	return sd.FileID
}

// Owner represents the distinguishable ID of the entity that
// holds the lock for FileID.
func (sd *LockDescriptor) Owner() string {
	return sd.UserID
}

// NewLockMapEntry returns an instance of a LockMapEntry
func NewLockMapEntry(owner string, timestamp time.Time) *LockMapEntry {
	return &LockMapEntry{
		owner:     owner,
		timestamp: timestamp,
	}
}

// NewSimpleConfig returns an instance of the SimpleConfig.
func NewSimpleConfig(IPAddr, PortAddr string) *SimpleConfig {
	return &SimpleConfig{
		IPAddr:   IPAddr,
		PortAddr: PortAddr,
	}
}

// NewLockDescriptor returns an instance of the LockDescriptor.
func NewLockDescriptor(FileID, UserID string) *LockDescriptor {
	return &LockDescriptor{
		FileID: FileID,
		UserID: UserID,
	}
}

// NewObjectDescriptor returns an instance of the ObjectDescriptor.
func NewObjectDescriptor(ObjectID string) *ObjectDescriptor {
	return &ObjectDescriptor{
		ObjectID: ObjectID,
	}
}

// NewSimpleLockService creates and returns a new lock service ready to use.
func NewSimpleLockService(log zerolog.Logger) *SimpleLockService {
	safeLockMap := &SafeLockMap{
		LockMap: make(map[string]*LockMapEntry),
	}
	return &SimpleLockService{
		log:     log,
		lockMap: safeLockMap,
	}
}

// getCurrentDuration returns the duration between the current time
// and the time at which a lock was required
func getCurrentDuration(timestamp time.Time) time.Duration {
	return time.Now().Sub(timestamp)
}

// Acquire function lets a client acquire a lock on an object.
func (ls *SimpleLockService) Acquire(sd Descriptors) error {
	ls.lockMap.Mutex.Lock()

	// If the lock is not present in the LockMap or
	// the lock has expired, then allow the acquisition
	if _, ok := ls.lockMap.LockMap[sd.ID()]; !ok || getCurrentDuration(ls.lockMap.LockMap[sd.ID()].timestamp) > ls.lockMap.LeaseDuration {
		ls.lockMap.LockMap[sd.ID()] = NewLockMapEntry(sd.Owner(), time.Now())
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Str("owner", ls.lockMap.LockMap[sd.ID()].owner).
			Time("timestamp", ls.lockMap.LockMap[sd.ID()].timestamp).
			Msg("locked")
		return nil
	}
	ls.lockMap.Mutex.Unlock()
	ls.
		log.
		Debug().
		Str("descriptor", sd.ID()).
		Msg("can't acquire, already been acquired")
	return ErrFileacquired

}

// Release lets a client to release a lock on an object.
// TODO: Prevent a lock from being released if it has expired.
func (ls *SimpleLockService) Release(sd Descriptors) error {
	ls.lockMap.Mutex.Lock()
	// Only the entity that posseses the lock for this object
	// is allowed to release the lock
	if ls.lockMap.LockMap[sd.ID()].owner == sd.Owner() {
		delete(ls.lockMap.LockMap, sd.ID())
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Str("owner", sd.Owner()).
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

// CheckAcquired returns true if the file is Acquired.
// It also returns the owner of the file.
func (ls *SimpleLockService) CheckAcquired(sd Descriptors) (string, bool) {
	ls.lockMap.Mutex.Lock()
	id := sd.ID()
	if entry, ok := ls.lockMap.LockMap[id]; ok {
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", id).
			Msg("checkacquire success")
		return entry.owner, true
	}
	ls.
		log.
		Debug().
		Str("descriptor", id).
		Msg("check acquire failure")
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
