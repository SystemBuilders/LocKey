package lockservice

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
)

type fsm RaftStore

type fsmSnapshot struct {
	lockMap map[string]string
}

func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	fmt.Println("reached apply")
	switch c.Op {
	case "acquire":
		return f.applyAcquire(c.Key, c.Value)

	case "release":
		return f.applyRelease(c.Key, c.Value)

	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))

	}
}

func (f *fsm) applyAcquire(lock, owner string) interface{} {
	fmt.Println("reached applyAcquire")
	desc := NewSimpleDescriptor(lock, owner)

	err := f.ls.Acquire(desc)
	if err != nil {
		return err
	}
	return nil
}

func (f *fsm) applyRelease(lock, owner string) interface{} {
	desc := NewSimpleDescriptor(lock, owner)

	err := f.ls.Release(desc)
	if err != nil {
		return err
	}
	return nil
}

// Snapshot returns a snapshot of the key-value store. We wrap
// the things we need in fsmSnapshot and then send that over to Persist.
// Persist encodes the needed data from fsmsnapshot and transport it to
// Restore where the necessary data is replicated into the finite state machine.
// This allows the consensus algorithm to truncate the replicated log.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.ls.lockMap.Mutex.Lock()
	defer f.ls.lockMap.Mutex.Unlock()

	return &fsmSnapshot{lockMap: f.ls.lockMap.LockMap}, nil
}

// Restores the lockMap to a previous state
func (f *fsm) Restore(lockMap io.ReadCloser) error {
	lockMapSnapshot := make(map[string]string)
	if err := json.NewDecoder(lockMap).Decode(&lockMapSnapshot); err != nil {
		return err
	}

	// Set the state from snapshot. No need to use mutex lock according
	// to Hasicorp doc
	f.ls.lockMap.LockMap = lockMapSnapshot

	return nil
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data
		b, err := json.Marshal(f.lockMap)
		if err != nil {
			return err
		}

		// Write data to sink
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink
		if err := sink.Close(); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (f *fsmSnapshot) Release() {}
