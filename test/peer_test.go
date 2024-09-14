package test

import (
	. "account/ledger"
	. "account/peer"
	. "account/transaction"
	"testing"
)

func TestPeer(t *testing.T) {
	ledger := MakeLedger()
	peer := new(Peer)
	peer.ID = "peer1"
	peer.Address = "localhost"
	peer.Connect("localhost", 8080)

	// test flood message
	peer.FloodMessage("hello")

	// test flood transaction
	tx := new(Transaction)
	tx.ID = "tx1"
	tx.From = "alice"
	tx.To = "bob"
	tx.Amount = 10
	peer.FloodTransaction(tx)

	// test ledger transaction
	ledger.Accounts["alice"] = 100
	ledger.Accounts["bob"] = 0
	ledger.Transaction(tx)
	if ledger.Accounts["alice"] != 90 {
		t.Errorf("ledger transaction failed")
	}
	if ledger.Accounts["bob"] != 10 {
		t.Errorf("ledger transaction failed")
	}
}
