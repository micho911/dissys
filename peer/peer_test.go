package peer

import (
	"fmt"
	"testing"
	"time"
)

// Test function
func TestConnection(t *testing.T) {
	// Create peers
	fmt.Println("Testing connections...")
	peerList := createPeers("1", "2", "3", "4")
	peer1, peer2, peer3, peer4 := peerList[0], peerList[1], peerList[2], peerList[3]

	// Start peer1
	startPeer(t, peer1, "localhost:0", 500*time.Millisecond)
	checkPeer(t, peer1, 1, 0, "Peer1 after starting")

	// Start peer2 and connect to peer1
	startPeer(t, peer2, peer1.Address, 500*time.Millisecond)
	printPeers(peer2, peer1)
	checkPeer(t, peer2, 2, 1, "Peer2 after connecting to Peer1")
	checkPeer(t, peer1, 2, 0, "Peer1 after Peer2 connected")

	// Start peer3 and connect to peer2
	startPeer(t, peer3, peer2.Address, 500*time.Millisecond)
	printPeers(peer3, peer2, peer1)
	checkPeer(t, peer3, 3, 2, "Peer3 after connecting to Peer2")
	checkPeer(t, peer2, 3, 1, "Peer2 after Peer3 connected")
	checkPeer(t, peer1, 3, 0, "Peer1 after Peer3 connected")

	// Start peer4 and connect to peer2
	startPeer(t, peer4, peer2.Address, 500*time.Millisecond)
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
	fmt.Println("\nTesting flood transactions...")
	// Create peers
	ledgers := []*Ledger{}
	for range 5 {
		ledgers = append(ledgers, createLedgerWithAccounts("alice", "bob", "amin", "bus", "gang"))
	}
	peers := createPeersWithLedgers(ledgers)
	peer1, peer2, peer3, peer4, peer5 := peers[0], peers[1], peers[2], peers[3], peers[4]

	startPeer(t, peer1, "localhost:0", 500*time.Millisecond)
	startPeer(t, peer2, peer1.Address, 500*time.Millisecond)
	startPeer(t, peer3, peer2.Address, 500*time.Millisecond)
	startPeer(t, peer4, peer2.Address, 500*time.Millisecond)
	startPeer(t, peer5, peer2.Address, 500*time.Millisecond)

	// Create a transaction
	t1 := &Transaction{
		ID:     "1",
		From:   "alice",
		To:     "bob",
		Amount: 10,
	}

	peer1.FloodTransaction(t1)
	time.Sleep(500 * time.Millisecond)

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
	time.Sleep(500 * time.Millisecond)
	//Check that peer 1 now has updated ledger after peer2 flooded
	if peer1.Ledger.Accounts["alice"] != 20 {
		fmt.Printf("Alice has %d in account", peer1.Ledger.Accounts["alice"])
		t.Fatalf("Expected peer %s's ledger to have 20 in Alice's account", peer1.Id)

	}
	if peer1.Ledger.Accounts["amin"] != -30 {
		t.Fatalf("Expected peer %s's ledger to have -30 in Bob's account", peer1.Id)
	}

	t3 := &Transaction{
		ID:     "3",
		From:   "amin",
		To:     "bob",
		Amount: 15,
	}
	peer3.FloodTransaction(t3)
	time.Sleep(500 * time.Millisecond)
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
	}
}

func TestFloodMultipleTransactions(t *testing.T) {
	fmt.Println("\n**************Integration test**************\n (*5 accounts on 10 peers sending 10 transactions each*)")
	// Create 5 accounts: account1, account2, account3, account4, account5
	accounts := []string{"account1", "account2", "account3", "account4", "account5"}

	// Create ledgers for the peers with the accounts
	ledgers := []*Ledger{}
	for i := 0; i < 10; i++ { // Create 10 peers
		ledgers = append(ledgers, createLedgerWithAccounts(accounts...))
	}
	peers := createPeersWithLedgers(ledgers)

	// Start peers in a network where each peer connects to the previous one
	startPeer(t, peers[0], "localhost:0", 500*time.Millisecond)
	for i := 1; i < len(peers); i++ {
		startPeer(t, peers[i], peers[i-1].Address, 500*time.Millisecond)
	}

	// Each peer sends 10 transactions involving the 5 accounts
	for _, peer := range peers {
		go func(p *Peer) {
			for j := 0; j < 10; j++ { // Send 10 transactions from each peer
				from := accounts[j%5] // Cycle through accounts
				to := accounts[(j+1)%5]
				amount := 10 * (j + 1) // Vary the transaction amount

				txn := &Transaction{
					ID:     fmt.Sprintf("txn-%s-%d", p.Id, j),
					From:   from,
					To:     to,
					Amount: amount,
				}

				// Flood transaction to the network
				p.FloodTransaction(txn)
			}
		}(peer)
	}

	// Give time for transactions to propagate
	time.Sleep(500 * time.Millisecond)

	// Check that all peers have the same ledger state
	for _, account := range accounts {
		expectedBalance := peers[0].Ledger.Accounts[account]
		for _, peer := range peers {
			if peer.Ledger.Accounts[account] != expectedBalance {
				t.Fatalf("Mismatch: Peer %s has %d in %s, expected %d", peer.Id, peer.Ledger.Accounts[account], account, expectedBalance)
			}
		}
	}

	t.Log("All peers have consistent ledger state across all accounts")
}
