package peer

import (
	"fmt"
	"transaction"
)

func (p *Peer) UpdateLedger(tx *transaction.Transaction, reply *bool) error {
	p.Ledger.Transaction(tx)
	*reply = true
	return nil
}

func (p *Peer) SignedUpdateLedger(tx *transaction.SignedTransaction, reply *bool) error {
	p.Ledger.SignedTransaction(tx)
	*reply = true
	return nil
}

func (p *Peer) GetPeers(args struct{}, reply *[]string) error {
	*reply = make([]string, 0, len(p.Peers))
	for addr := range p.Peers {
		*reply = append(*reply, addr)
	}
	return nil
}

func (p *Peer) JoinMessage(addr string, reply *bool) error {
	if _, exists := p.Peers[addr]; !exists {
		p.Peers[addr] = nil // Add peer with no connection
		fmt.Printf("Peer %s received join message from %s\n", p.Id, addr)
		*reply = true
	}
	*reply = true
	return nil
}
