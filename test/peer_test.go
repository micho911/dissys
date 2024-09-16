package test

import (
	. "account/ledger"
	. "account/peer"
	"time"

	// . "account/transaction"
	"testing"
)

func TestPeer(t *testing.T) {
	// currently test only that the peer can be created and two peers can connect to each other
	peer1 := Peer{ID: "1", Ledger: MakeLedger(), Peers: []string{}}
	go peer1.Connect("localhost:0")
	time.Sleep(2 * time.Second)
	if len(peer1.Peers) != 1 {
		t.Errorf("Peer1 should have 1 peer in its list of peers, but has %d", len(peer1.Peers))
	}
	peer2 := Peer{ID: "2", Ledger: MakeLedger(), Peers: []string{}}
	go peer2.Connect(peer1.Adress)
	time.Sleep(1 * time.Second)
	// test that when peer 2 connects to peer 1, peer 1 has peer 2 in its list of peers and peer 2 has itself in its list of peers
	if len(peer1.Peers) != 2 {
		t.Errorf("Peer1 should have 2 peers in its list of peers, but has %d", len(peer1.Peers))
	}
	if len(peer2.Peers) != 1 {
		t.Errorf("Peer2 should have 1 peer in its list of peers, but has %d", len(peer2.Peers))
	}

	peer3 := Peer{ID: "3", Ledger: MakeLedger(), Peers: []string{}}
	go peer3.Connect(peer1.Adress)
	time.Sleep(1 * time.Second)

	peer4 := Peer{ID: "4", Ledger: MakeLedger(), Peers: []string{}}
	go peer4.Connect(peer3.Adress)
	time.Sleep(1 * time.Second)

	// test that peer1 has itself in its list of peers

	peer3.FloodMessage("Hello")
}
