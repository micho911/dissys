package peer

import (
	"fmt"
	"testing"
	"time"
)

type idLedgerPair struct {
	id     string
	ledger *Ledger
}

// Test function
func TestConnection(t *testing.T) {
	// Create peers
	peerList := createPeers("1", "2", "3", "4")
	peer1, peer2, peer3, peer4 := peerList[0], peerList[1], peerList[2], peerList[3]

	// Start peer1
	startPeer(t, peer1, "localhost:0", 1000*time.Millisecond)
	checkPeer(t, peer1, 1, 0, "Peer1 after starting")

	// Start peer2 and connect to peer1
	startPeer(t, peer2, peer1.Address, 3000*time.Millisecond)
	printPeers(peer2, peer1)
	checkPeer(t, peer2, 2, 1, "Peer2 after connecting to Peer1")
	checkPeer(t, peer1, 2, 0, "Peer1 after Peer2 connected")

	// Start peer3 and connect to peer2
	startPeer(t, peer3, peer2.Address, 5000*time.Millisecond)
	printPeers(peer3, peer2, peer1)
	checkPeer(t, peer3, 3, 2, "Peer3 after connecting to Peer2")
	checkPeer(t, peer2, 3, 1, "Peer2 after Peer3 connected")
	checkPeer(t, peer1, 3, 0, "Peer1 after Peer3 connected")

	// Start peer4 and connect to peer2
	startPeer(t, peer4, peer2.Address, 1000*time.Millisecond)
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
	ids := []string{"1", "2", "3", "4", "5"}
	ledgers := []*Ledger{}
	for range ids {
		ledgers = append(ledgers, createLedgerWithAccounts("alice", "bob", "amin", "bus", "gang"))
	}
	peers := createPeersWithLedgers(createIdLedgerPairs(ids, ledgers)...)
	peer1, peer2, peer3, peer4, peer5 := peers[0], peers[1], peers[2], peers[3], peers[4]

	startPeer(t, peer1, "localhost:0", 1000*time.Millisecond)
	startPeer(t, peer2, peer1.Address, 1000*time.Millisecond)
	startPeer(t, peer3, peer2.Address, 1000*time.Millisecond)
	startPeer(t, peer4, peer2.Address, 1000*time.Millisecond)
	startPeer(t, peer5, peer2.Address, 1000*time.Millisecond)

	// Create a transaction
	transaction := &Transaction{
		ID:     "1",
		From:   "alice",
		To:     "bob",
		Amount: 10,
	}

	peer1.FloodTransaction(transaction)
	time.Sleep(2000 * time.Millisecond)
	fmt.Println("Ledgers after transaction")
	for _, p := range peers {
		fmt.Printf("Peer%s: %v\n", p.Id, p.Ledger.Accounts)
	}

	// Check that all ledgers have the same balance
	for _, p := range peers {
		for account, balance := range p.Ledger.Accounts {
			if account == "alice" {
				if balance != -10 {
					t.Fatalf("Expected alice to have -10, got %d (from peer %s)", balance, p.Id)

				}
			} else if account == "bob" {
				if balance != 10 {
					t.Fatalf("Expected bob to have 10, got %d (from peer %s)", balance, p.Id)
				}
			} else {
				if balance != 0 {
					t.Fatalf("Expected %s to have 0, got %d (from peer %s)", account, balance, p.Id)
				}
			}
		}
	}

}

func createLedgerWithAccounts(accounts ...string) *Ledger {
	l := MakeLedger()
	for _, account := range accounts {
		l.Accounts[account] = 0
	}
	return l
}

func createPeersWithLedgers(idLedgerPairs ...idLedgerPair) []*Peer {
	peers := make([]*Peer, len(idLedgerPairs))
	for i, pair := range idLedgerPairs {
		peers[i] = &Peer{Id: pair.id, Ledger: pair.ledger}
	}
	return peers
}

func createIdLedgerPairs(ids []string, ledgers []*Ledger) []idLedgerPair {
	pairs := make([]idLedgerPair, len(ids))
	for i, id := range ids {
		pairs[i] = idLedgerPair{id, ledgers[i]}
	}
	return pairs
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
