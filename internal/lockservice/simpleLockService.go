package lockservice

import (
	"fmt"
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
	FileID string `json:"fileID"`
	UserID string `json:"userID"`
}

// LockCheckRequest is an instance of a lock check request.
type LockCheckRequest struct {
	FileID string `json:"fileID"`
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
var _ Object = (*ObjectDescriptor)(nil)

// ObjectDescriptor describes the object that is subjected to
// lock operations.
type ObjectDescriptor struct {
	ObjectID string
}

// ID returns the ID related to the object.
func (od *ObjectDescriptor) ID() string {
	return od.ObjectID
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
func NewSimpleLockService(log zerolog.Logger, leaseDuration time.Duration) *SimpleLockService {
	safeLockMap := &SafeLockMap{
		LockMap: make(map[string]*LockMapEntry),
	}
	safeLockMap.LeaseDuration = leaseDuration
	return &SimpleLockService{
		log:     log,
		lockMap: safeLockMap,
	}
}
func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	ms := d / time.Microsecond
	return fmt.Sprintf("%02d:%02d", h, ms)
}

// hasLeaseExpired returns true if the lease for a lock has expired and
// false if the lease is still valid
func hasLeaseExpired(timestamp time.Time, lease time.Duration) bool {
	if time.Now().Sub(timestamp) > lease {
		return true
	}
	return false
}

// Acquire function lets a client acquire a lock on an object.
// This lock is valid for a fixed duration that is set in the SafeLockMap.LeaseDuration
// field. Beyond this duration, the lock has expired and the entity that owned the lock
// for this period can no longer release it. The lock is open for acquisition after it
// has expired.
func (ls *SimpleLockService) Acquire(sd Descriptors) error {
	ls.lockMap.Mutex.Lock()

	timestamp := ls.lockMap.LockMap[sd.ID()].timestamp
	duration := ls.lockMap.LeaseDuration
	// If the lock is not present in the LockMap or
	// the lock has expired, then allow the acquisition
	if _, ok := ls.lockMap.LockMap[sd.ID()]; !ok || hasLeaseExpired(timestamp, duration) {
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
	// Since the lock is already acquired, return an error
	ls.
		log.
		Debug().
		Str("descriptor", sd.ID()).
		Msg("can't acquire, already been acquired")
	return ErrFileacquired

}

// Release lets a client to release a lock on an object.
func (ls *SimpleLockService) Release(sd Descriptors) error {
	ls.lockMap.Mutex.Lock()
	timestamp := ls.lockMap.LockMap[sd.ID()].timestamp
	duration := ls.lockMap.LeaseDuration
	// Only the entity that posseses the lock for this object
	// is allowed to release the lock
	if _, ok := ls.lockMap.LockMap[sd.ID()]; !ok {
		// trying to release an unacquired lock
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("can't release, hasn't been acquired")
		ls.lockMap.Mutex.Unlock()
		return ErrCantReleaseFile

	} else if hasLeaseExpired(timestamp, duration) {
		// lease expired
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("can't release, lease of lock has expired")
		ls.lockMap.Mutex.Unlock()
		return ErrCantReleaseFile

	} else if ls.lockMap.LockMap[sd.ID()].owner == sd.Owner() {
		// conditions satisfied, lock is released
		delete(ls.lockMap.LockMap, sd.ID())
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Str("owner", sd.Owner()).
			Msg("released")
		ls.lockMap.Mutex.Unlock()
		return nil
	} else {
		// trying to release a lock that you don't own
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
