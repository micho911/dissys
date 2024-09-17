package main

import (
	L "account/ledger"
	P "account/peer"
	"time"
)

func main() {
	// currently test only that the peer can be created and two peers can connect to each other
	peer1 := P.Peer{ID: "1", Ledger: L.MakeLedger(), Adress: "localhost:0"}
	go peer1.Connect("localhost:0")
	time.Sleep(1 * time.Second)
	peer2 := P.Peer{ID: "2", Ledger: L.MakeLedger()}
	go peer2.Connect(peer1.Adress)
	time.Sleep(1 * time.Second)
	peer3 := P.Peer{ID: "3", Ledger: L.MakeLedger()}
	go peer3.Connect(peer1.Adress)
	time.Sleep(1 * time.Second)
	peer4 := P.Peer{ID: "4", Ledger: L.MakeLedger()}
	go peer4.Connect(peer2.Adress)
	time.Sleep(1 * time.Second)

	// peer3.FloodMessage("Hello")

	// tx := &transaction.Transaction{
	// 	ID:     "tx1",
	// 	From:   "Alice",
	// 	To:     "Bob",
	// 	Amount: 10,
	// }

	// myPeer.FloodTransaction(tx)

	// fmt.Println("Ledger after transaction:", myPeer.Ledger.Accounts)
}
