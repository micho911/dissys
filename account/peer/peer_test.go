package peer

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

// Test function
func TestConnection(t *testing.T) {
	// Create peers
	peerList := createPeers("1", "2", "3", "4")
	peer1, peer2, peer3, peer4 := peerList[0], peerList[1], peerList[2], peerList[3]

	// Start peer1
	startPeer(t, peer1, "localhost:0", 3000*time.Millisecond)
	checkPeer(t, peer1, 1, 0, "Peer1 after starting")

	// Start peer2 and connect to peer1
	startPeer(t, peer2, peer1.Address, 3000*time.Millisecond)
	printPeers(peer2, peer1)
	checkPeer(t, peer2, 2, 1, "Peer2 after connecting to Peer1")
	checkPeer(t, peer1, 2, 0, "Peer1 after Peer2 connected")

	// Start peer3 and connect to peer2
	startPeer(t, peer3, peer2.Address, 3000*time.Millisecond)
	printPeers(peer3, peer2, peer1)
	checkPeer(t, peer3, 3, 2, "Peer3 after connecting to Peer2")
	checkPeer(t, peer2, 3, 1, "Peer2 after Peer3 connected")
	checkPeer(t, peer1, 3, 0, "Peer1 after Peer3 connected")

	// Start peer4 and connect to peer2
	startPeer(t, peer4, peer2.Address, 3000*time.Millisecond)
	printPeers(peer4)
	printPeers(peer3)
	printPeers(peer2)
	printPeers(peer1)
	checkPeer(t, peer4, 4, 3, "Peer4 after connecting to Peer2")
	checkPeer(t, peer3, 4, 2, "Peer3 after Peer4 connected")
	checkPeer(t, peer2, 4, 1, "Peer2 after Peer4 connected")
	checkPeer(t, peer1, 4, 0, "Peer1 after Peer4 connected")
}

func TestFloodTransaction(t *testing.T) {
	// Create peers
	ledgers := []*Ledger{}
	for range 5 {
		ledgers = append(ledgers, createLedgerWithAccounts("alice", "bob", "amin", "bus", "gang"))
	}
	peers := createPeersWithLedgers(ledgers)
	peer1, peer2, peer3, peer4, peer5 := peers[0], peers[1], peers[2], peers[3], peers[4]

	startPeer(t, peer1, "localhost:0", 1000*time.Millisecond)
	startPeer(t, peer2, peer1.Address, 1000*time.Millisecond)
	startPeer(t, peer3, peer2.Address, 1000*time.Millisecond)
	startPeer(t, peer4, peer2.Address, 1000*time.Millisecond)
	startPeer(t, peer5, peer2.Address, 1000*time.Millisecond)

	// Create a transaction
	t1 := &Transaction{
		ID:     "1",
		From:   "alice",
		To:     "bob",
		Amount: 10,
	}

	peer1.FloodTransaction(t1)
	time.Sleep(2000 * time.Millisecond)

	//Check that peer 1 now has updated ledger
	if peer1.Ledger.Accounts["alice"] != -10 {
		t.Fatalf("Expected peer %s's ledger to have -10 in Alice's account", peer1.Id)
	}
	if peer1.Ledger.Accounts["bob"] != 10 {
		t.Fatalf("Expected peer %s's ledger to have 10 in Bob's account", peer1.Id)
	}

	t2 := &Transaction{
		ID:     "2",
		From:   "amin",
		To:     "alice",
		Amount: 30,
	}
	peer2.FloodTransaction(t2)
	time.Sleep(3000 * time.Millisecond)
	//Check that peer 1 now has updated ledger after peer2 flooded
	if peer1.Ledger.Accounts["alice"] != 20 {
		fmt.Printf("Alice has %d in account", peer1.Ledger.Accounts["alice"])
		t.Fatalf("Expected peer %s's ledger to have 20 in Alice's account", peer1.Id)

	}
	if peer1.Ledger.Accounts["amin"] != -30 {
		t.Fatalf("Expected peer %s's ledger to have -30 in Bob's account", peer1.Id)
	}

	/*t3 := &Transaction{
		ID:     "3",
		From:   "amin",
		To:     "bob",
		Amount: 15,
	}
	peer3.FloodTransaction(t3)
	time.Sleep(3000 * time.Millisecond)
	//Check that peer 1 now has updated ledgers after peer3 flooded
	if peer1.Ledger.Accounts["amin"] != -45 {
		fmt.Printf("Amin has %d", peer1.Ledger.Accounts["amin"])
		t.Fatalf("Expected peer %s's ledger to have -45 in Amin's account", peer1.Id)

	}
	if peer1.Ledger.Accounts["bob"] != 25 {
		t.Fatalf("Expected peer %s's ledger to have -25 in Bob's account", peer1.Id)
	}
	//Check that peer 1 now has updated ledgers after peer3 flooded
	if peer2.Ledger.Accounts["amin"] != -45 {
		t.Fatalf("Expected peer %s's ledger to have -45 in Amin's account", peer2.Id)

	}
	if peer2.Ledger.Accounts["bob"] != 25 {
		t.Fatalf("Expected peer %s's ledger to have -25 in Bob's account", peer2.Id)
	} */
}

func createLedgerWithAccounts(accounts ...string) *Ledger {
	l := MakeLedger()
	for _, account := range accounts {
		l.Accounts[account] = 0
	}
	return l
}

func createPeersWithLedgers(ledgers []*Ledger) []*Peer {
	peers := make([]*Peer, len(ledgers))
	for i, ledger := range ledgers {
		peers[i] = &Peer{Id: strconv.Itoa(i + 1), Ledger: ledger}
	}
	return peers
}

// Helper function to create peers with given IDs
func createPeers(ids ...string) []*Peer {
	peers := make([]*Peer, len(ids))
	for i, id := range ids {
		peers[i] = &Peer{Id: id, Ledger: MakeLedger()}
	}
	return peers
}

// Helper function to start a peer and connect it to an address
func startPeer(t *testing.T, p *Peer, connectAddr string, delay time.Duration) {
	go p.Connect(connectAddr)
	time.Sleep(delay)
}

func connectionCount(p *Peer) int {
	count := 0
	for _, client := range p.Peers {
		if client != nil {
			count++
		}
	}
	return count
}

// Helper function to check the state of a peer
func checkPeer(t *testing.T, p *Peer, expectedPeers int, expectedConnections int, msg string) {
	if len(p.Peers) != expectedPeers {
		t.Fatalf("%s: expected %d peers, got %d", msg, expectedPeers, len(p.Peers))
	}
	if connectionCount(p) != expectedConnections {
		t.Fatalf("%s: expected %d connections, got %d", msg, expectedConnections, connectionCount(p))
	}
}

// Helper function to print peer maps for debugging
func printPeers(peers ...*Peer) {
	for _, p := range peers {
		fmt.Printf("Peer%s.peers: %v\n", p.Id, p.Peers)
	}
}
