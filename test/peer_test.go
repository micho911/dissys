package test

import (
	. "account/ledger"
	. "account/peer"
	"fmt"
	"time"

	// . "account/transaction"
	"testing"
)

func testAddress(t *testing.T) {
	peer1 := Peer{ID: "1", Ledger: MakeLedger()}
	go peer1.Connect("localhost:0")
	time.Sleep(2 * time.Second)
	if peer1.Address == "" {
		t.Errorf("Peer address should not be empty")
	}
	// test other cases
	peer2 := Peer{ID: "2", Ledger: MakeLedger()}
	go peer2.Connect(peer1.Address)
	time.Sleep(2 * time.Second)
	if peer2.Address == "" {
		t.Errorf("Peer address should not be empty")
	}

}

func TestPeer(t *testing.T) {
	// currently test only that the peer can be created and two peers can connect to each other
	peer1 := Peer{ID: "1", Ledger: MakeLedger()}
	go peer1.Connect("localhost:0")
	time.Sleep(2 * time.Second)

	if len(peer1.Peers.Members()) != 1 {
		t.Errorf("Peer1 should have 1 peer in its list of peers, but has %d", len(peer1.Peers.Members()))
	}
	peer2 := Peer{ID: "2", Ledger: MakeLedger()}

	go peer2.Connect(peer1.Address)
	time.Sleep(2 * time.Second)

	// test that when peer 2 connects to peer 1, peer 1 has peer 2 in its list of peers and peer 2 has itself in its list of peers
	if len(peer1.Peers.Members()) != 2 {
		t.Errorf("Peer1 should have 2 peers in its list of peers, but has %d", len(peer1.Peers.Members()))
	}
	if len(peer2.Peers.Members()) != 2 {
		t.Errorf("Peer2 should have 1 peer in its list of peers, but has %d", len(peer2.Peers.Members()))
	}

	// print peer Set
	fmt.Println("Peer1 Peers:", peer1.Peers.Members())
	fmt.Println("Peer2 Peers:", peer2.Peers.Members())

	peer3 := Peer{ID: "3", Ledger: MakeLedger()}
	go peer3.Connect(peer1.Address)
	time.Sleep(2 * time.Second)

	if len(peer3.Peers.Members()) != 3 {
		t.Errorf("Peer3 should have 3 peer in its list of peers, but has %d", len(peer3.Peers.Members()))
	}
	if len(peer1.Peers.Members()) != 3 {
		t.Errorf("Peer1 should have 3 peers in its list of peers, but has %d", len(peer1.Peers.Members()))
	}
	if len(peer2.Peers.Members()) != 3 {
		t.Errorf("Peer2 should have 3 peers in its list of peers, but has %d", len(peer2.Peers.Members()))
	}
	fmt.Println("Peer1 Peers:", peer1.Peers.Members())
	fmt.Println("Peer2 Peers:", peer2.Peers.Members())
	fmt.Println("Peer3 Peers:", peer3.Peers.Members())

	peer4 := Peer{ID: "4", Ledger: MakeLedger()}
	fmt.Println("Peer 3 address", peer3.Address)
	go peer4.Connect(peer3.Address)
	time.Sleep(1 * time.Second)
	fmt.Println("Peer4 Peers:", peer4.Peers.Members())
	if len(peer4.Peers.Members()) != 4 {
		t.Errorf("Peer4 should have 4 peer in its list of peers, but has %d", len(peer4.Peers.Members()))
	}
	fmt.Println("Peer1 Peers:", peer1.Peers.Members())
	fmt.Println("Peer2 Peers:", peer2.Peers.Members())
	fmt.Println("Peer3 Peers:", peer3.Peers.Members())
	fmt.Println("Peer4 Peers:", peer4.Peers.Members())
	// test that peer1 has itself in its list of peers

	// peer3.FloodMessage("Hello")
}
