package peer

import (
	"ledger"
	"log"
)

func (p *Peer) FloodSignedTransaction(tx *ledger.SignedTransaction) {
	p.Ledger.SignedTransaction(tx)
	for addr := range p.Peers {
		senderAddr := p.Address
		if addr == senderAddr {
			continue
		}
		if p.Peers[addr] == nil {
			p.connectToPeer(addr)
		}
		client := p.Peers[addr]
		if client == nil {
			log.Printf("Peer %s could not connect to %s", p.Id, addr)
			continue
		}
		var reply bool
		method := "Peer.SignedUpdateLedger"
		success := false
		for i := 0; i < 3; i++ {
			err := client.Call(method, tx, &reply)
			if err == nil && reply {
				success = true
				break
			}
			log.Println("Failed to send transaction, retrying:", err)
		}
		if !success {
			log.Printf("Failed to send transaction to %s after retries", addr)
		}
	}
}
