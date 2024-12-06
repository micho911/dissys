package peer

import (
	"crypto/sha256"
	"encrypt"
	"fmt"
	"ledger"
	"log"
	"nsc"
)

func (p *Peer) UpdateLedger(tx *ledger.SignedTransaction, reply *bool) error {
	p.Ledger.SignedTransaction(tx)
	*reply = true
	return nil
}

func (p *Peer) GetBlockHistory(args struct{}, reply *[]*nsc.Block) error {
	p.ReceivedBlocks.lock.Lock()
	defer p.ReceivedBlocks.lock.Unlock()
	*reply = append(*reply, *p.ReceivedBlocks.Blocks...)
	return nil
}

func (p *Peer) GetPeers(args struct{}, reply *[]string) error {
	p.Peers.lock.Lock()
	defer p.Peers.lock.Unlock()
	*reply = make([]string, 0, len(p.Peers.Clients))
	for addr := range p.Peers.Clients {
		*reply = append(*reply, addr)
	}
	return nil
}

func (p *Peer) JoinMessage(addr string, reply *bool) error {
	p.Peers.lock.Lock()
	defer p.Peers.lock.Unlock()
	fmt.Printf("Peer %s received join message from %s\n", encrypt.EncodeKey(p.Vk)[5:12], addr)
	if _, exists := p.Peers.Clients[addr]; !exists {
		p.Peers.Clients[addr] = nil
	}
	*reply = true
	return nil
}

func (p *Peer) ProcessBlock(block *nsc.Block) bool {
	p.Tree.Lock.Lock()
	defer p.Tree.Lock.Unlock()

	if !block.Verify(p.Tree) {
		fmt.Printf("Peer %s rejected a block\n", encrypt.EncodeKey(p.Vk)[5:12])
		return false
	}
	fmt.Printf("Peer %s verified a block\n", encrypt.EncodeKey(p.Vk)[5:12])

	oldLeader := p.Tree.FindLeadingNode()

	fmt.Printf("Peer %s is adding block to tree\n", encrypt.EncodeKey(p.Vk)[5:12])
	addedToTree := p.Tree.AddBlock(block)
	if !addedToTree {
		fmt.Printf("Peer %s registered an orphan\n", encrypt.EncodeKey(p.Vk)[5:12])
	}

	newLeader := p.Tree.FindLeadingNode()
	oldLeaderSerializedBlock, _ := nsc.SerializeBlock(oldLeader.Block)
	oldLeaderHash := sha256.Sum256(oldLeaderSerializedBlock)

	if newLeader.Block.Ptr != oldLeaderHash {
		p.Rollback(oldLeader)
		p.catchUp()
		return true
	}
	p.executeBlock(block)
	return true
}

type Args struct {
	Block      *nsc.Block
	Recipients *[]string
}

func (p *Peer) ReceiveBlock(args *Args, reply *bool) error {
	p.ReceivedBlocks.lock.Lock()
	defer p.ReceivedBlocks.lock.Unlock() // Ensure lock is released no matter what happens

	fmt.Printf("Peer %s received block\n", encrypt.EncodeKey(p.Vk)[5:12])

	if p.ReceivedBlocks.Blocks != nil {
		for _, block := range *p.ReceivedBlocks.Blocks {
			for _, blockTx := range block.Transactions {
				for _, newTx := range args.Block.Transactions {
					if blockTx.Signature == newTx.Signature {
						*reply = true
						return nil
					}
				}
			}
		}
	}

	success := p.ProcessBlock(args.Block)

	if success {
		*p.ReceivedBlocks.Blocks = append(*p.ReceivedBlocks.Blocks, args.Block)
		fmt.Printf("Block processed and added to received blocks\n")

		*args.Recipients = append(*args.Recipients, p.Address)

		go p.FloodBlock(args)
	}

	*reply = true
	return nil
}

func (p *Peer) FloodBlock(args *Args) error {
	for addr := range p.Peers.Clients {

		if args.Recipients == nil {
			fmt.Print("Recipients list is nil\n")
		}

		skip := false
		for _, recipient := range *args.Recipients {
			if addr == recipient {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		if p.Peers.Clients[addr] == nil {
			p.connectToPeer(addr)
		}

		client := p.Peers.Clients[addr]
		if client == nil {
			log.Printf("Peer %s could not connect to %s\n", encrypt.EncodeKey(p.Vk), addr)
			continue
		}

		var received bool
		method := "Peer.ReceiveBlock"
		success := false
		for i := 0; i < 3; i++ {
			err := client.Call(method, args, &received)
			if err == nil && received {
				success = true
				break
			}
			log.Println("Failed to flood block, retrying:", err)
		}

		if !success {
			log.Printf("Failed to flood block to %s after retries", addr)
		}
	}

	return nil
}
