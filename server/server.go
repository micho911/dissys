package main

import (
	L "account/ledger"
	P "account/peer"
	"bufio"
	"fmt"
	"os"
)

func main() {
	// currently test only that the peer can be created and two peers can connect to each other
	// ask the user for the adress to connect to
	fmt.Println("Enter the adress to connect to:")
	reader := bufio.NewReader(os.Stdin)
	address, _ := reader.ReadString('\n')
	// trim the newline character and whitespace
	address = address[:len(address)-1]

	// ask user for name of peer
	fmt.Println("Enter the name of the peer:")
	name, _ := reader.ReadString('\n')
	ID := name
	ID = ID[:len(ID)-1]

	peer := P.Peer{ID: ID, Ledger: L.MakeLedger(), Adress: address}
	peer.Connect(address)

	// tx := &transaction.Transaction{
	// 	ID:     "tx1",
	// 	From:   "Alice",
	// 	To:     "Bob",
	// 	Amount: 10,
	// }

	// myPeer.FloodTransaction(tx)

	// fmt.Println("Ledger after transaction:", myPeer.Ledger.Accounts)
}
