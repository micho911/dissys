package main

import (
	"fmt"
	"peer"
	"time"
	"transaction"
)

func main() {
	fmt.Println("\n************** Running Peer to Peer Network **************")
	fmt.Printf("(* 5 accounts on 10 peers sending 10 transactions each *)\n\n")

	accounts := []string{"account1", "account2", "account3", "account4", "account5"}

	ledgers := []*transaction.Ledger{}
	for i := 0; i < 10; i++ {
		ledgers = append(ledgers, createLedgerWithAccounts(accounts...))
	}
	peers := createPeersWithLedgers(ledgers)

	startPeer(peers[0], "localhost:0", 500*time.Millisecond)
	for i := 1; i < len(peers); i++ {
		startPeer(peers[i], peers[i-1].Address, 500*time.Millisecond)
	}

	for _, p := range peers {
		go func(p *peer.Peer) {
			for j := 0; j < 10; j++ {
				from := accounts[j%5]
				to := accounts[(j+1)%5]
				amount := 10 * (j + 1)

				txn := &transaction.Transaction{
					ID:     fmt.Sprintf("txn-%s-%d", p.Id, j),
					From:   from,
					To:     to,
					Amount: amount,
				}

				p.FloodTransaction(txn)
			}
		}(p)
	}

	time.Sleep(2 * time.Second)

	fmt.Printf("************** Ledgers **************\n\n")
	printLedgers(peers)
}
