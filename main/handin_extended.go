package main

import (
	"fmt"
	"ledger"
	"peer"
	"sort"
	"time"
)

func createLedgerWithAccounts(accounts ...string) *ledger.Ledger {
	l := ledger.MakeLedger()
	for _, account := range accounts {
		l.Accounts[account] = 0
	}
	return l
}

func createPeersWithLedgers(ledgers []*ledger.Ledger) []*peer.Peer {
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
		accounts := make([]string, 0, len(p.Ledger.Accounts))
		for account := range p.Ledger.Accounts {
			accounts = append(accounts, account)
		}
		sort.Strings(accounts)
		for _, account := range accounts {
			balance := p.Ledger.Accounts[account]
			accountShort := account
			if len(account) > 10 {
				accountShort = account[5:15] + "..."
			}
			fmt.Printf(" - %s: %d\n", accountShort, balance)
		}
		fmt.Println()
	}
}
