package peer

import (
	T "account/transaction"
)

type Peer struct {
	ID      string
	Address string
}

func (p *Peer) Connect(addr string, port int) {
	// connect peer to network
}

func (p *Peer) FloodMessage(msg string) {
	// flood message to network
}

func (p *Peer) FloodTransaction(tx *T.Transaction) {
	// flood transaction to network
}
