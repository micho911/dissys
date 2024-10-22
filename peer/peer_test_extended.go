package peer

import (
	"fmt"
	"strconv"
	"testing"
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

func createPeersWithLedgers(ledgers []*transaction.Ledger) []*Peer {
	peers := make([]*Peer, len(ledgers))
	for i, ledger := range ledgers {
		peers[i] = &Peer{Id: strconv.Itoa(i + 1), Ledger: ledger}
	}
	return peers
}

func createPeers(ids ...string) []*Peer {
	peers := make([]*Peer, len(ids))
	for i, id := range ids {
		peers[i] = &Peer{Id: id, Ledger: transaction.MakeLedger()}
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

func checkPeer(t *testing.T, p *Peer, expectedPeers int, expectedConnections int, msg string) {
	if len(p.Peers) != expectedPeers {
		t.Fatalf("%s: expected %d peers, got %d", msg, expectedPeers, len(p.Peers))
	}
	if connectionCount(p) != expectedConnections {
		t.Fatalf("%s: expected %d connections, got %d", msg, expectedConnections, connectionCount(p))
	}
}

func printPeers(peers ...*Peer) {
	for _, p := range peers {
		fmt.Printf("Peer%s.peers: %v\n", p.Id, p.Peers)
	}
}
