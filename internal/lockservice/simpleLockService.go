package lockservice

import (
	"sync"

	"github.com/rs/zerolog"
)

// SafeLockMap is the lockserver's data structure
type SafeLockMap struct {
	LockMap map[string]bool
	Mutex   sync.Mutex
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
}

// ID represents the distinguishable ID of the descriptor.
func (sd *SimpleDescriptor) ID() string {
	return sd.FileID
}

// NewSimpleLockService creates and returns a new lock service ready to use.
func NewSimpleLockService(log zerolog.Logger) *SimpleLockService {
	safeLockMap := &SafeLockMap{
		LockMap: make(map[string]bool),
	}
	return &SimpleLockService{
		log:     log,
		lockMap: safeLockMap,
	}
}

// Acquire function lets a client acquire a lock on an object.
func (ls *SimpleLockService) Acquire(sd Descriptors) error {
	ls.lockMap.Mutex.Lock()
	if ls.lockMap.LockMap[sd.ID()] {
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("can't acquire, already been acquired")
		return ErrFileAcquired
	}
	ls.lockMap.LockMap[sd.ID()] = true
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
	if ls.lockMap.LockMap[sd.ID()] {
		delete(ls.lockMap.LockMap, sd.ID())
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("released")
		ls.lockMap.Mutex.Unlock()
		return nil
	}
	ls.
		log.
		Debug().
		Str("descriptor", sd.ID()).
		Msg("can't release, hasn't been acquired")
	ls.lockMap.Mutex.Unlock()
	return ErrCantReleaseFile
}

// CheckAcquired returns true if the file is acquired
func (ls *SimpleLockService) CheckAcquired(sd Descriptors) bool {
	ls.lockMap.Mutex.Lock()
	if ls.lockMap.LockMap[sd.ID()] {
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("checkAcquire success")
		return true
	}
	ls.
		log.
		Debug().
		Str("descriptor", sd.ID()).
		Msg("check Acquire failure")
	ls.lockMap.Mutex.Unlock()
	return false
}

// CheckReleased returns true if the file is released
func (ls *SimpleLockService) CheckReleased(sd Descriptors) bool {
	ls.lockMap.Mutex.Lock()
	if ls.lockMap.LockMap[sd.ID()] {
		ls.lockMap.Mutex.Unlock()
		ls.
			log.
			Debug().
			Str("descriptor", sd.ID()).
			Msg("checkRelease failure")
		return false
	}
	ls.
		log.
		Debug().
		Str("descriptor", sd.ID()).
		Msg("checkRelease success")
	ls.lockMap.Mutex.Unlock()
	return true
}
