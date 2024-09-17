package test

import (
	. "account/ledger"
	. "account/peer"
	"fmt"
	"time"

	// . "account/transaction"
	"testing"
)

func TestPeer(t *testing.T) {
	// currently test only that the peer can be created and two peers can connect to each other
	peer1 := Peer{ID: "1", Ledger: MakeLedger()}
	go peer1.Connect("localhost:0")
	time.Sleep(2 * time.Second)
	if len(peer1.Peers.Peers) != 1 {
		t.Errorf("Peer1 should have 1 peer in its list of peers, but has %d", len(peer1.Peers.Peers))
	}
	peer2 := Peer{ID: "2", Ledger: MakeLedger()}
	fmt.Println("Peer1 adress:", peer1.Adress)
	printPeerList(peer1.Peers)
	printPeerList(peer2.Peers)
	go peer2.Connect("localhost:0")

	time.Sleep(1 * time.Second)

	// test that when peer 2 connects to peer 1, peer 1 has peer 2 in its list of peers and peer 2 has itself in its list of peers
	// if len(peer1.Peers.Peers) != 2 {
	// 	t.Errorf("Peer1 should have 2 peers in its list of peers, but has %d", len(peer1.Peers.Peers))
	// }
	if len(peer2.Peers.Peers) != 1 {
		t.Errorf("Peer2 should have 1 peer in its list of peers, but has %d", len(peer2.Peers.Peers))
	}

	peer3 := Peer{ID: "3", Ledger: MakeLedger()}
	go peer3.Connect(peer1.Adress)
	time.Sleep(1 * time.Second)
	if len(peer3.Peers.Peers) != 1 {
		t.Errorf("Peer3 should have 1 peer in its list of peers, but has %d", len(peer3.Peers.Peers))
	}

	peer4 := Peer{ID: "4", Ledger: MakeLedger()}
	go peer4.Connect(peer3.Adress)
	time.Sleep(1 * time.Second)

	// test that peer1 has itself in its list of peers

	// peer3.FloodMessage("Hello")
}

func printPeerList(peers PeerList) {
	for _, peer := range peers.Peers {
		fmt.Println(peer)
	}
}
