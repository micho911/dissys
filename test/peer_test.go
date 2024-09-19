package peer

import (
	"fmt"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {

	// Create two peers
	peer1 := &Peer{Id: "Peer1"}
	peer2 := &Peer{Id: "Peer2"}
	peer3 := &Peer{Id: "Peer3"}
	peer4 := &Peer{Id: "Peer4"}

	go peer1.Connect("localhost:0")
	time.Sleep(1000 * time.Millisecond)
	if len(peer1.Peers) != 1 {
		t.Fatal("Peer1 should itself in peer list")
	}
	if peer1.ConnectionCount() != 0 {
		t.Fatal("Peer1 should have no outgoing connections")
	}

	go peer2.Connect(peer1.Address)
	time.Sleep(3000 * time.Millisecond)
	fmt.Println("Peer2.peers: ", peer2.Peers)
	fmt.Println("Peer1.peers: ", peer1.Peers)
	if len(peer2.Peers) != 2 {
		t.Fatal("Peer2 should have itself and peer 1 in its peer list")
	}
	if peer2.ConnectionCount() != 1 {
		t.Fatal("Peer2 should one outgoing connection")
	}
	if len(peer1.Peers) != 2 {
		t.Fatal("Peer1 should receive join message from peer 2")
	}
	if peer1.ConnectionCount() != 0 {
		t.Fatal("Peer1 should have no new connections")
	}

	go peer3.Connect(peer2.Address)
	time.Sleep(5000 * time.Millisecond)
	fmt.Println("Peer3.peers: ", peer3.Peers)
	fmt.Println("Peer2.peers: ", peer2.Peers)
	fmt.Println("Peer1.peers: ", peer1.Peers)
	if len(peer3.Peers) != 3 {
		t.Fatal("Peer3 should have retrieved peer list from Peer2 and add itself")
	}
	if peer3.ConnectionCount() != 2 {
		t.Fatal("Peer3 should two outgoing connection -> peer1 and peer2 ")
	}
	if len(peer2.Peers) != 3 {
		t.Fatal("Peer2 should receive join message from peer 3")
	}
	if peer2.ConnectionCount() != 1 {
		t.Fatal("Peer2 should have no new connections")
	}
	if len(peer1.Peers) != 3 {
		t.Fatal("Peer1 should receive join message from peer 3")
	}
	if peer1.ConnectionCount() != 0 {
		t.Fatal("Peer1 should have no new connections")
	}

	go peer4.Connect(peer2.Address)
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("Peer4.peers: ", peer4.Peers)
	if len(peer4.Peers) != 4 {
		t.Fatal("Peer4 should have retrieved peer list from peer 2")
	}
	if peer4.ConnectionCount() != 3 {
		t.Fatal("Peer4 should three outgoing connection -> peer1 and peer2 ")
	}

	/* fmt.Println("Peer3.peers: ", peer3.Peers)
	if len(peer3.Peers) != 4 {
		t.Fatal("Peer3 should have received join message from peer4")
	}
	if peer3.ConnectionCount() != 2 {
		t.Fatal("Peer3 should have no new connections")
	} */

	/* peer2.Connect(peer3.Address)
	time.Sleep(1000 * time.Millisecond)
	if len(peer2.Peers) != 4 {
		t.Fatal("Peer2 should have retrieved peer list from Peer1")
	}

	peer3.Connect(peer2.Address)
	time.Sleep(1000 * time.Millisecond)
	if len(peer3.Peers) != 4 {
		t.Fatal("Peer2 should have retrieved peer list from Peer1")
	}  */
}
