package peer

import (
	"fmt"
	"ledger"
	"nsc"
)

func (p *Peer) UpdateLedger(tx *ledger.SignedTransaction, reply *bool) error {
	p.Ledger.SignedTransaction(*tx)
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

func (p *Peer) ProcessBlock(block *nsc.Block, sender string, reply *bool) error {
	if !block.Verify(p.Tree) {
		*reply = true
		return nil
	}

	//Add block to tree and
	//handle rollback if we switch branch..

	if sender != p.Address {
		p.FloodBlock(block)
	}

	*reply = true
	return nil
}

func (p *Peer) AddToTransactionBuffer(tx *ledger.SignedTransaction, sender string, reply *bool) error {
	p.transactionBuffer.lock.Lock()
	defer p.transactionBuffer.lock.Unlock()

	for _, existingTx := range p.transactionBuffer.transactions {
		if existingTx.Signature == tx.Signature {
			*reply = true
			return nil
		}
	}

	p.transactionBuffer.transactions = append(p.transactionBuffer.transactions, tx)

	if sender != p.Address {
		p.FloodTransaction(tx)
	}

	*reply = true
	return nil
}
