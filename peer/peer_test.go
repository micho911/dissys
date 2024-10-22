package peer

import (
	"encoding/base64"
	"encrypt"
	"fmt"
	"ledger"
	"strconv"
	"testing"
	"time"
)

type KeyPair struct {
	PublicKey  encrypt.Key
	PrivateKey encrypt.Key
}

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
	accounts := []string{"alice", "bob", "amin", "bus", "gang"}
	keyPairs := make(map[string]KeyPair)
	encodedAccounts := make(map[string]string)

	for _, account := range accounts {
		pk, sk := encrypt.KeyGen(2048)
		keyPairs[account] = KeyPair{PublicKey: pk, PrivateKey: sk}
		encodedAccounts[account] = encrypt.EncodeKey(pk)
	}

	accountKeys := make([]string, 0, len(encodedAccounts))
	for _, encodedKey := range encodedAccounts {
		accountKeys = append(accountKeys, encodedKey)
	}

	// Create peers
	ledgers := []*ledger.Ledger{}
	for range 5 {
		ledgers = append(ledgers, createLedgerWithAccounts(accountKeys...))
	}
	peers := createPeersWithLedgers(ledgers)
	peer1, peer2, peer3, peer4, peer5 := peers[0], peers[1], peers[2], peers[3], peers[4]

	startPeer(t, peer1, "localhost:0", 500*time.Millisecond)
	startPeer(t, peer2, peer1.Address, 500*time.Millisecond)
	startPeer(t, peer3, peer2.Address, 500*time.Millisecond)
	startPeer(t, peer4, peer2.Address, 500*time.Millisecond)
	startPeer(t, peer5, peer2.Address, 500*time.Millisecond)

	// Create a transaction
	txnID := "1"
	fromAccount := encodedAccounts["alice"]
	toAccount := encodedAccounts["bob"]
	amount := 10
	msg := txnID + fromAccount + toAccount + strconv.Itoa(amount)
	signatureBytes := encrypt.Sign([]byte(msg), keyPairs["alice"].PrivateKey)
	signatureStr := base64.StdEncoding.EncodeToString(signatureBytes)

	t1 := &ledger.SignedTransaction{
		ID:        txnID,
		From:      fromAccount,
		To:        toAccount,
		Amount:    amount,
		Signature: signatureStr,
	}

	peer1.FloodTransaction(t1)
	time.Sleep(500 * time.Millisecond)

	//Check that peer 1 now has updated ledger
	if peer1.Ledger.Accounts[encodedAccounts["alice"]] != -10 {

		t.Fatalf("Expected peer %s's ledger to have -10 in Alice's account", peer1.Id)
	}
	if peer1.Ledger.Accounts[encodedAccounts["bob"]] != 10 {
		t.Fatalf("Expected peer %s's ledger to have 10 in Bob's account", peer1.Id)
	}

	txnID = "2"
	fromAccount = encodedAccounts["amin"]
	toAccount = encodedAccounts["alice"]
	amount = 30
	msg = txnID + fromAccount + toAccount + strconv.Itoa(amount)
	signatureBytes = encrypt.Sign([]byte(msg), keyPairs["amin"].PrivateKey)
	signatureStr = base64.StdEncoding.EncodeToString(signatureBytes)

	t2 := &ledger.SignedTransaction{
		ID:        txnID,
		From:      fromAccount,
		To:        toAccount,
		Amount:    amount,
		Signature: signatureStr,
	}

	peer2.FloodTransaction(t2)
	time.Sleep(500 * time.Millisecond)
	//Check that peer 1 now has updated ledger after peer2 flooded
	if peer1.Ledger.Accounts[encodedAccounts["alice"]] != 20 {
		fmt.Printf("Alice has %d in account", peer1.Ledger.Accounts["alice"])
		t.Fatalf("Expected peer %s's ledger to have 20 in Alice's account", peer1.Id)

	}
	if peer1.Ledger.Accounts[encodedAccounts["amin"]] != -30 {
		t.Fatalf("Expected peer %s's ledger to have -30 in Bob's account", peer1.Id)
	}

	txnID = "3"
	fromAccount = encodedAccounts["amin"]
	toAccount = encodedAccounts["bob"]
	amount = 15
	msg = txnID + fromAccount + toAccount + strconv.Itoa(amount)
	signatureBytes = encrypt.Sign([]byte(msg), keyPairs["amin"].PrivateKey)
	signatureStr = base64.StdEncoding.EncodeToString(signatureBytes)

	t3 := &ledger.SignedTransaction{
		ID:        txnID,
		From:      fromAccount,
		To:        toAccount,
		Amount:    amount,
		Signature: signatureStr,
	}
	peer3.FloodTransaction(t3)
	time.Sleep(500 * time.Millisecond)
	if peer1.Ledger.Accounts[encodedAccounts["amin"]] != -45 {
		t.Fatalf("Expected peer %s's ledger to have -45 in Amin's account", peer1.Id)

	}
	if peer1.Ledger.Accounts[encodedAccounts["bob"]] != 25 {
		t.Fatalf("Expected peer %s's ledger to have -25 in Bob's account", peer1.Id)
	}
	if peer2.Ledger.Accounts[encodedAccounts["amin"]] != -45 {
		t.Fatalf("Expected peer %s's ledger to have -45 in Amin's account", peer2.Id)

	}
	if peer2.Ledger.Accounts[encodedAccounts["bob"]] != 25 {
		t.Fatalf("Expected peer %s's ledger to have -25 in Bob's account", peer2.Id)
	}
}

func TestInvalidTransactions(t *testing.T) {
	fmt.Println("\nTesting invalid transactions...")

	// Setup accounts and keys
	accounts := []string{"alice", "bob", "charlie"}
	keyPairs := make(map[string]KeyPair)
	encodedAccounts := make(map[string]string)

	for _, account := range accounts {
		pk, sk := encrypt.KeyGen(2048)
		keyPairs[account] = KeyPair{PublicKey: pk, PrivateKey: sk}
		encodedAccounts[account] = encrypt.EncodeKey(pk)
	}

	accountKeys := []string{encodedAccounts["alice"], encodedAccounts["bob"], encodedAccounts["charlie"]}

	// Create ledgers with accounts
	ledgers := []*ledger.Ledger{}
	for i := 0; i < 3; i++ {
		ledgers = append(ledgers, createLedgerWithAccounts(accountKeys...))
	}
	peers := createPeersWithLedgers(ledgers)
	peer1, peer2, peer3 := peers[0], peers[1], peers[2]

	startPeer(t, peer1, "localhost:0", 500*time.Millisecond)
	startPeer(t, peer2, peer1.Address, 500*time.Millisecond)
	startPeer(t, peer3, peer1.Address, 500*time.Millisecond)

	// Initial balances
	initialBalance := 100
	for _, peer := range peers {
		peer.Ledger.Accounts[encodedAccounts["alice"]] = initialBalance
		peer.Ledger.Accounts[encodedAccounts["bob"]] = initialBalance
		peer.Ledger.Accounts[encodedAccounts["charlie"]] = initialBalance
	}

	txnID := "invalid-1"
	fromAccount := encodedAccounts["alice"]
	toAccount := encodedAccounts["bob"]
	amount := 50
	msg := txnID + fromAccount + toAccount + strconv.Itoa(amount)
	signatureBytes := encrypt.Sign([]byte(msg), keyPairs["bob"].PrivateKey) // Wrong key
	signatureStr := base64.StdEncoding.EncodeToString(signatureBytes)

	invalidTxn1 := &ledger.SignedTransaction{
		ID:        txnID,
		From:      fromAccount,
		To:        toAccount,
		Amount:    amount,
		Signature: signatureStr,
	}

	peer1.FloodTransaction(invalidTxn1)
	time.Sleep(500 * time.Millisecond)

	for _, peer := range peers {
		if peer.Ledger.Accounts[fromAccount] != initialBalance {
			t.Fatalf("Peer %s: Expected Alice's balance to remain %d, got %d", peer.Id, initialBalance, peer.Ledger.Accounts[fromAccount])
		}
		if peer.Ledger.Accounts[toAccount] != initialBalance {
			t.Fatalf("Peer %s: Expected Bob's balance to remain %d, got %d", peer.Id, initialBalance, peer.Ledger.Accounts[toAccount])
		}
	}

	txnID = "invalid-2"
	fromAccount = encodedAccounts["alice"]
	toAccount = encodedAccounts["charlie"]
	amount = 30
	msg = txnID + fromAccount + toAccount + strconv.Itoa(amount)
	signatureBytes = encrypt.Sign([]byte(msg), keyPairs["alice"].PrivateKey)
	signatureBytes[0] ^= 0xFF // Tamper with the signature
	signatureStr = base64.StdEncoding.EncodeToString(signatureBytes)

	invalidTxn2 := &ledger.SignedTransaction{
		ID:        txnID,
		From:      fromAccount,
		To:        toAccount,
		Amount:    amount,
		Signature: signatureStr,
	}

	peer2.FloodTransaction(invalidTxn2)
	time.Sleep(500 * time.Millisecond)

	for _, peer := range peers {
		if peer.Ledger.Accounts[fromAccount] != initialBalance {
			t.Fatalf("Peer %s: Expected Alice's balance to remain %d, got %d", peer.Id, initialBalance, peer.Ledger.Accounts[fromAccount])
		}
		if peer.Ledger.Accounts[toAccount] != initialBalance {
			t.Fatalf("Peer %s: Expected Charlie's balance to remain %d, got %d", peer.Id, initialBalance, peer.Ledger.Accounts[toAccount])
		}
	}

	txnID = "invalid-3"
	fromAccount = encodedAccounts["bob"]
	toAccount = encodedAccounts["alice"]
	amount = 20
	wrongMsg := txnID + fromAccount + toAccount + strconv.Itoa(amount+10)       // Modify the amount
	signatureBytes = encrypt.Sign([]byte(wrongMsg), keyPairs["bob"].PrivateKey) // Sign wrong message
	signatureStr = base64.StdEncoding.EncodeToString(signatureBytes)

	invalidTxn3 := &ledger.SignedTransaction{
		ID:        txnID,
		From:      fromAccount,
		To:        toAccount,
		Amount:    amount,
		Signature: signatureStr,
	}

	peer3.FloodTransaction(invalidTxn3)
	time.Sleep(500 * time.Millisecond)

	// Check that balances did not change
	for _, peer := range peers {
		if peer.Ledger.Accounts[fromAccount] != initialBalance {
			t.Fatalf("Peer %s: Expected Bob's balance to remain %d, got %d", peer.Id, initialBalance, peer.Ledger.Accounts[fromAccount])
		}
		if peer.Ledger.Accounts[toAccount] != initialBalance {
			t.Fatalf("Peer %s: Expected Alice's balance to remain %d, got %d", peer.Id, initialBalance, peer.Ledger.Accounts[toAccount])
		}
	}

	// 4. Valid Transaction for control
	txnID = "valid-1"
	fromAccount = encodedAccounts["alice"]
	toAccount = encodedAccounts["bob"]
	amount = 40
	msg = txnID + fromAccount + toAccount + strconv.Itoa(amount)
	signatureBytes = encrypt.Sign([]byte(msg), keyPairs["alice"].PrivateKey)
	signatureStr = base64.StdEncoding.EncodeToString(signatureBytes)

	validTxn := &ledger.SignedTransaction{
		ID:        txnID,
		From:      fromAccount,
		To:        toAccount,
		Amount:    amount,
		Signature: signatureStr,
	}

	peer1.FloodTransaction(validTxn)
	time.Sleep(500 * time.Millisecond)

	// Check that balances updated correctly
	expectedAliceBalance := initialBalance - amount
	expectedBobBalance := initialBalance + amount
	for _, peer := range peers {
		if peer.Ledger.Accounts[fromAccount] != expectedAliceBalance {
			t.Fatalf("Peer %s: Expected Alice's balance to be %d, got %d", peer.Id, expectedAliceBalance, peer.Ledger.Accounts[fromAccount])
		}
		if peer.Ledger.Accounts[toAccount] != expectedBobBalance {
			t.Fatalf("Peer %s: Expected Bob's balance to be %d, got %d", peer.Id, expectedBobBalance, peer.Ledger.Accounts[toAccount])
		}
	}
}

func TestFloodMultipleTransactions(t *testing.T) {
	fmt.Println("\n**************Integration test**************\n (*5 accounts on 10 peers sending 10 transactions each*)")
	accounts := []string{"account1", "account2", "account3", "account4", "account5"}
	keyPairs := make(map[string]KeyPair)
	encodedAccounts := make(map[string]string)

	for _, account := range accounts {
		pk, sk := encrypt.KeyGen(1024)
		keyPairs[account] = KeyPair{PublicKey: pk, PrivateKey: sk}
		encodedAccounts[account] = encrypt.EncodeKey(pk)
	}

	accountKeys := make([]string, 0, len(encodedAccounts))
	for _, encodedKey := range encodedAccounts {
		accountKeys = append(accountKeys, encodedKey)
	}

	ledgers := []*ledger.Ledger{}
	for i := 0; i < 10; i++ {
		ledgers = append(ledgers, createLedgerWithAccounts(accountKeys...))
	}
	peers := createPeersWithLedgers(ledgers)

	startPeer(t, peers[0], "localhost:0", 500*time.Millisecond)
	for i := 1; i < len(peers); i++ {
		startPeer(t, peers[i], peers[i-1].Address, 500*time.Millisecond)
	}

	// Each peer sends 10 transactions involving the 5 accounts
	for _, peer := range peers {
		go func(p *Peer) {
			for j := 0; j < 10; j++ {
				fromName := accounts[j%5]
				toName := accounts[(j+1)%5]
				amount := 10 * (j + 1)

				fromAccount := encodedAccounts[fromName]
				toAccount := encodedAccounts[toName]
				txnID := fmt.Sprintf("txn-%s-%d", p.Id, j)
				msg := txnID + fromAccount + toAccount + strconv.Itoa(amount)
				signatureBytes := encrypt.Sign([]byte(msg), keyPairs[fromName].PrivateKey)
				signatureStr := base64.StdEncoding.EncodeToString(signatureBytes)

				txn := &ledger.SignedTransaction{
					ID:        txnID,
					From:      fromAccount,
					To:        toAccount,
					Amount:    amount,
					Signature: signatureStr,
				}

				// Flood transaction to the network
				p.FloodTransaction(txn)
			}
		}(peer)
	}

	time.Sleep(500 * time.Millisecond)

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
