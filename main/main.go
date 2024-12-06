package main

import (
	"encoding/base64"
	"encrypt"
	"fmt"
	"ledger"
	"math/rand"
	"nsc"
	"peer"
	"sort"
	"strconv"
	"strings"
	"time"
)

func PrintBlockTree(node *nsc.Node, depth int) {
	if node == nil {
		return
	}

	// Print the block and its transactions
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%sBlock:\n", indent)
	fmt.Printf("%s  Draw Winner: %x\n", indent, encrypt.EncodeKey(node.Block.Draw.Vk)[:5]) // Print first few bytes of the Draw key for brevity
	fmt.Printf("%s  Transactions:\n", indent)
	for _, tx := range node.Block.Transactions {
		fmt.Printf("%s    Tx: %s, From: %s, To: %s, Amount: %d\n", indent, tx.ID, tx.From[:5], tx.To[:5], tx.Amount)
	}

	// Recursively print child nodes
	for _, child := range node.Children {
		PrintBlockTree(child, depth+1)
	}
}

func startPeer(p *peer.Peer, connectAddr string, delay time.Duration) {
	go p.Connect(connectAddr)
	time.Sleep(delay)
}

func printLedgers(peers []*peer.Peer) {
	for _, p := range peers {
		fmt.Printf("Peer %s's ledger:\n", encrypt.EncodeKey(p.Vk)[5:12])
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

func generateRandomTransaction(peers []*peer.Peer) *ledger.SignedTransaction {
	fromPeer := peers[rand.Intn(len(peers))]
	toPeer := peers[rand.Intn(len(peers))]

	// Ensure it's not the same peer
	for fromPeer == toPeer {
		toPeer = peers[rand.Intn(len(peers))]
	}

	from := encrypt.EncodeKey(fromPeer.Vk)
	to := encrypt.EncodeKey(toPeer.Vk)
	amount := rand.Intn(100) + 1 // Random amount between 1 and 100
	time := time.Now().Unix()
	// Create transaction ID and message
	txID := fmt.Sprintf("txn-%d", rand.Intn(1000))
	msg := txID + from + to + strconv.Itoa(amount) + strconv.FormatInt(time, 10)

	// Sign the transaction
	signature := encrypt.Sign([]byte(msg), fromPeer.Sk)

	// Create and return the signed transaction
	tx := &ledger.SignedTransaction{
		ID:        txID,
		From:      from,
		To:        to,
		Amount:    ledger.AU(amount),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Time:      time,
	}

	return tx
}

func SimulatePeerNetwork() {
	// Generate genesis wallets
	nsc.GenerateGenesisWallets()

	// Get the list of genesis wallets
	genesisWallets := nsc.GetGenesisWallets()

	// Create genesis peers
	var peers []*peer.Peer
	for i := 0; i < 10; i++ {
		peer := peer.MakePeer((*genesisWallets)[i])
		peers = append(peers, peer)
	}

	// n Start peers
	startPeer(peers[0], "1", 0*time.Millisecond)

	time.Sleep(1 * time.Second)

	for i := 1; i < 10; i++ {
		startPeer(peers[i], peers[i-1].Address, 300*time.Millisecond)
	}
	time.Sleep(2 * time.Second)
	for range 100 {
		peer := peers[rand.Intn(len(peers))]
		randomTx := generateRandomTransaction(peers)
		peer.MakeTransaction(randomTx)
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(10 * time.Second)

	// //Make new peeers
	// for i := 0; i < 5; i++ {
	// 	pk, sk := encrypt.KeyGen(2048)
	// 	peer := peer.MakePeer(nsc.Wallet{PublicKey: pk, PrivateKey: sk})
	// 	peers = append(peers, peer)
	// 	startPeer(peer, peers[rand.Intn(10)].Address, 100*time.Millisecond)
	// }

	// fmt.Printf("Updated peers %d", len(peers))

	// for range 100 {
	// 	peer := peers[rand.Intn(len(peers))]
	// 	randomTx := generateRandomTransaction(peers)
	// 	peer.MakeTransaction(randomTx)
	// 	time.Sleep(100 * time.Millisecond)
	// }

	//time.Sleep(20 * time.Second)

	printLedgers(peers)

	for _, peer := range peers {
		leader := peer.Tree.FindLeadingNode().Block.Draw.Vk
		fmt.Printf("Peer %s has leading block won by %s\n", encrypt.EncodeKey(peer.Vk)[5:12], encrypt.EncodeKey(leader)[5:12])
	}
	fmt.Printf("Updated peers %d", len(peers))

}

func main() {
	SimulatePeerNetwork()
}
