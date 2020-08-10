package lockservice

import (
	"fmt"
	"testing"
)

func TestRaft(t *testing.T) {
	raftLS, err := NewRaftServer(
		"test",
		"127.0.0.1:5000",
	)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		fmt.Println("no error!")
	}
	fmt.Printf("%v", raftLS)

}
