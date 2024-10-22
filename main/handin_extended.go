package main

import (
	"fmt"
	"peer"
	"sort"
	"time"
	"transaction"
)

func createLedgerWithAccounts(accounts ...string) *transaction.Ledger {
	l := transaction.MakeLedger()
	for _, account := range accounts {
		l.Accounts[account] = 0
	}
	return l
}

func createPeersWithLedgers(ledgers []*transaction.Ledger) []*peer.Peer {
	peers := make([]*peer.Peer, len(ledgers))
	for i, ledger := range ledgers {
		peers[i] = &peer.Peer{
			Id:     fmt.Sprintf("%d", i+1),
			Ledger: ledger,
		}
	}
	return peers
}

func startPeer(p *peer.Peer, connectAddr string, delay time.Duration) {
	go p.Connect(connectAddr)
	time.Sleep(delay)
}

func printLedgers(peers []*peer.Peer) {
	for _, p := range peers {
		fmt.Printf("Peer %s's ledger:\n", p.Id)
		// Extract account names and sort them
		accounts := make([]string, 0, len(p.Ledger.Accounts))
		for account := range p.Ledger.Accounts {
			accounts = append(accounts, account)
		}
		sort.Strings(accounts)
		// Print accounts in sorted order
		for _, account := range accounts {
			balance := p.Ledger.Accounts[account]
			fmt.Printf(" - %s: %d\n", account, balance)
		}
		fmt.Println()
	}
}
