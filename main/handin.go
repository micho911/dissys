package main

import (
	"encoding/base64"
	"encrypt"
	"fmt"
	"ledger"
	"peer"
	"strconv"
	"time"
)

type KeyPair struct {
	PublicKey  encrypt.Key
	PrivateKey encrypt.Key
}

func main() {
	fmt.Println("\n************** Running Peer to Peer Network **************")
	fmt.Printf("(* 5 accounts on 10 peers sending 10 transactions each *)\n\n")

	accounts := []string{"account1", "account2", "account3", "account4", "account5"}
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

	// Create ledgers for the peers with the accounts
	ledgers := []*ledger.Ledger{}
	for i := 0; i < 10; i++ { // Create 10 peers
		ledgers = append(ledgers, createLedgerWithAccounts(accountKeys...))
	}
	peers := createPeersWithLedgers(ledgers)

	startPeer(peers[0], "localhost:0", 500*time.Millisecond)
	for i := 1; i < len(peers); i++ {
		startPeer(peers[i], peers[i-1].Address, 500*time.Millisecond)
	}

	for _, p := range peers {
		go func(p *peer.Peer) {
			for j := 0; j < 10; j++ {
				fromName := accounts[j%5]
				toName := accounts[(j+1)%5]
				amount := 10 * (j + 1)

<<<<<<< main/handin.go
				fromAccount := encodedAccounts[fromName]
				toAccount := encodedAccounts[toName]
				txnID := fmt.Sprintf("txn-%s-%d", p.Id, j)
				msg := txnID + fromAccount + toAccount + strconv.Itoa(amount)
				signatureBytes := encrypt.Sign([]byte(msg), keyPairs[fromName].PrivateKey)
				signatureStr := base64.StdEncoding.EncodeToString(signatureBytes)

				txn := &peer.SignedTransaction{
					ID:        txnID,
					From:      fromAccount,
					To:        toAccount,
					Amount:    amount,
					Signature: signatureStr,
=======
				txn := &ledger.Transaction{
					ID:     fmt.Sprintf("txn-%s-%d", p.Id, j),
					From:   from,
					To:     to,
					Amount: amount,
>>>>>>> main/handin.go
				}

				p.FloodTransaction(txn)
			}
		}(p)
	}

	time.Sleep(2 * time.Second)

	fmt.Printf("************** Ledgers **************\n\n")
	printLedgers(peers)
}
