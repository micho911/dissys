package peer

import (
	"crypto/sha256"
	"encrypt"
	"fmt"
	"ledger"
	"log"
	"net"
	"net/rpc"
	"nsc"
	"strconv"
	"sync"
	"time"
)

type Peers struct {
	Clients map[string]*rpc.Client
	lock    sync.Mutex
}
type ReceivedBlocks struct {
	Blocks *[]*nsc.Block
	lock   sync.Mutex
}

type Peer struct {
	Address           string
	Vk                encrypt.Key
	Sk                encrypt.Key
	Peers             *Peers
	Ledger            *ledger.Ledger
	Tree              *nsc.Tree
	transactionBuffer *TransactionBuffer
	ReceivedBlocks    *ReceivedBlocks
}

type TransactionBuffer struct {
	transactions []*ledger.SignedTransaction
	lock         sync.Mutex
}

func MakePeer(wallet nsc.Wallet) *Peer {
	tree := nsc.NewTree()
	peer := &Peer{
		Ledger:            ledger.MakeLedger(),
		Sk:                wallet.PrivateKey,
		Vk:                wallet.PublicKey,
		transactionBuffer: &TransactionBuffer{},
		Tree:              tree,
		Peers:             &Peers{},
		ReceivedBlocks:    &ReceivedBlocks{Blocks: &[]*nsc.Block{}},
	}
	return peer
}

func (p *Peer) MakeTransaction(tx *ledger.SignedTransaction) {
	p.transactionBuffer.lock.Lock()
	defer p.transactionBuffer.lock.Unlock()
	fmt.Printf("Peer %s made transaction %s\n", encrypt.EncodeKey(p.Vk)[5:12], tx.ID)
	p.transactionBuffer.transactions = append(p.transactionBuffer.transactions, tx)
}

func (p *Peer) Connect(addr string) {
	p.Peers.Clients = map[string]*rpc.Client{}
	client, err := rpc.Dial("tcp", addr)
	if err != nil || client == nil {
		p.serve()
		p.Tree.Lock.Lock()
		p.catchUp()
		p.Tree.Lock.Unlock()
		p.Mine()
	} else {
		p.Peers.Clients[addr] = client
		fmt.Printf("Peer %s connected to %s\n", encrypt.EncodeKey(p.Vk)[5:12], addr)
		p.serve()
		p.requestPeers(client)
		p.requestBlockHistory()
		p.FloodJoinMessage()
		time.Tick(500 * time.Millisecond)
		p.Mine()
	}

}

func (p *Peer) Mine() {
	fmt.Printf("Peer %s is mining\n", encrypt.EncodeKey(p.Vk)[6:12])
	go p.joinLottery()
}

func (p *Peer) serve() {
	server := rpc.NewServer()
	err := server.RegisterName("Peer", p)
	if err != nil {
		log.Fatalf("Error registering RPC methods: %v", err)
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to listen on localhost:0: %v", err)
	}
	p.Address = l.Addr().String()
	p.Peers.lock.Lock()
	p.Peers.Clients[p.Address] = nil
	defer p.Peers.lock.Unlock()
	fmt.Printf("Peer %s serving at %s\n", encrypt.EncodeKey(p.Vk)[5:12], p.Address)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("accept error:", err)
				continue
			}
			go server.ServeConn(conn)
		}
	}()
}

func (p *Peer) joinLottery() {
	previousSlot, _ := p.Tree.GetCurrentSlot()
	ticker := time.Tick(1 * time.Second)
	for {
		currentSlot, currentTime := p.Tree.GetCurrentSlot()
		if currentSlot > previousSlot {
			p.ProcessNewSlot(currentSlot, currentTime)
			previousSlot = currentSlot
		}
		<-ticker
	}
}

func (p *Peer) MakeBlock(draw nsc.Draw) *nsc.Block {
	p.transactionBuffer.lock.Lock()
	defer p.transactionBuffer.lock.Unlock()

	p.Tree.Lock.Lock()
	defer p.Tree.Lock.Unlock()

	var transactions []*ledger.SignedTransaction
	if len(p.transactionBuffer.transactions) <= p.Tree.BlockSize {
		transactions = p.transactionBuffer.transactions
	} else {
		transactions = p.transactionBuffer.transactions[:p.Tree.BlockSize]
	}

	if len(transactions) == 0 {
		return nil
	}

	serializedTransactions, _ := ledger.SerializeTransactions(transactions)
	sig := encrypt.Sign(serializedTransactions, p.Sk)

	prevBlock := p.Tree.FindLeadingNode().Block
	serializedPrevBlock, _ := nsc.SerializeBlock(prevBlock)
	prevBlockHash := sha256.Sum256(serializedPrevBlock)

	newBlock := &nsc.Block{
		Transactions: transactions,
		Draw:         draw,
		Ptr:          prevBlockHash,
		Sig:          sig,
	}
	fmt.Printf("Peer %s mined a block with %s transactions\n", encrypt.EncodeKey(p.Vk)[5:12], strconv.Itoa(len(newBlock.Transactions)))

	p.transactionBuffer.transactions = []*ledger.SignedTransaction{}

	return newBlock
}

func (p *Peer) ProcessNewSlot(slot int64, time int64) {
	draw := p.draw(slot, time)
	if draw.IsWinner(p.Tree) {
		fmt.Printf("Peer %s won slot %s\n", encrypt.EncodeKey(p.Vk)[5:12], strconv.FormatInt(slot, 10))
		block := p.MakeBlock(draw)
		if block == nil {
			return
		}
		args := &Args{Block: block, Recipients: &[]string{}}
		var reply bool
		go p.ReceiveBlock(args, &reply)
		fmt.Printf("Peer %s flooded block\n", encrypt.EncodeKey(p.Vk)[5:12])
	}
}

func (p *Peer) Rollback(oldLeader *nsc.Node) error {
	fmt.Printf("Peer %s is performing a rollback\n", encrypt.EncodeKey(p.Vk)[5:12])
	current := oldLeader
	for current != nil {
		transactions := current.Block.Transactions
		for i := len(transactions) - 1; i >= 0; i-- {
			signedTx := transactions[i]
			tx := &ledger.Transaction{ID: signedTx.ID, From: signedTx.From, To: signedTx.To, Amount: signedTx.Amount}
			p.Ledger.UndoTransaction(tx)
		}
		current = current.Parent
	}
	return nil
}

func (p *Peer) catchUp() error {
	leader := p.Tree.FindLeadingNode()
	var blocksInPath []*nsc.Block

	current := leader
	for current != nil {
		blocksInPath = append(blocksInPath, current.Block)
		current = current.Parent
	}

	for i := len(blocksInPath) - 1; i >= 0; i-- {
		block := blocksInPath[i]
		transactions := block.Transactions
		for _, tx := range transactions {
			p.Ledger.SignedTransaction(tx)
		}
	}

	fmt.Printf("Peer %s caught up\n", encrypt.EncodeKey(p.Vk)[5:12])
	return nil
}

func (p *Peer) executeBlock(block *nsc.Block) {
	transactions := block.Transactions
	fmt.Printf("Peer %s executing block:\n", encrypt.EncodeKey(p.Vk)[5:12])
	for _, tx := range transactions {
		p.Ledger.SignedTransaction(tx) // Remember to check balances
		fmt.Printf("\t%s\n", tx.ID)
	}
	prize := len(transactions) + 10
	recipient := encrypt.EncodeKey(block.Draw.Vk)
	p.Ledger.Accounts[recipient] += prize
}

func (p *Peer) draw(slot int64, time int64) nsc.Draw {
	str := "lottery"
	slotStr := strconv.FormatInt(slot, 10)
	seed := p.Tree.Seed
	seedStr := strconv.Itoa(seed)
	drawStr := str + slotStr + seedStr
	drawSign := encrypt.Sign([]byte(drawStr), p.Sk)
	draw := nsc.MakeDraw(p.Vk, time, slot, drawSign)
	fmt.Printf("Peer %s drew at slot %s\n", encrypt.EncodeKey(p.Vk)[5:12], strconv.FormatInt(slot, 10))
	return draw
}

func (p *Peer) connectToPeer(addr string) *rpc.Client {
	client, err := rpc.Dial("tcp", addr)
	if err != nil || client == nil {
		fmt.Printf("Peer %s could not connect to %s\n", encrypt.EncodeKey(p.Vk), addr)
		return nil
	}
	p.Peers.Clients[addr] = client
	return client
}

func (p *Peer) AddReceivedBlocksFromPeer(newBlocks []*nsc.Block) {
	p.ReceivedBlocks.lock.Lock()
	defer p.ReceivedBlocks.lock.Unlock()

	for _, newBlock := range newBlocks {
		if !p.hasBlock(newBlock) {
			*p.ReceivedBlocks.Blocks = append(*p.ReceivedBlocks.Blocks, newBlock)
		}
	}
}

func (p *Peer) hasBlock(block *nsc.Block) bool {
	for _, existingBlock := range *p.ReceivedBlocks.Blocks {
		if block == existingBlock {
			return true
		}
	}
	return false
}

func (p *Peer) requestBlockHistory() {
	p.Tree.Lock.Lock()
	defer p.Tree.Lock.Unlock()
	for addr := range p.Peers.Clients {
		senderAddr := p.Address
		if addr == senderAddr {
			continue
		}

		if p.Peers.Clients[addr] == nil {
			p.connectToPeer(addr)
		}
		client := p.Peers.Clients[addr]
		var reply *[]*nsc.Block
		method := "Peer.GetBlockHistory"
		success := false
		for i := 0; i < 3; i++ {
			err := client.Call(method, struct{}{}, &reply)
			if err == nil && reply != nil {
				p.AddReceivedBlocksFromPeer(*reply)
				success = true
				break
			}
			fmt.Printf("Err: %s", err)
		}
		if !success {
			log.Printf("Failed to request block history to %s after retries", addr)
		}
	}
	for _, block := range *p.ReceivedBlocks.Blocks {
		p.Tree.AddBlock(block)
	}
	p.catchUp()
}

func (p *Peer) requestPeers(client *rpc.Client) {
	fmt.Printf("Peer %s is requesting peers\n", encrypt.EncodeKey(p.Vk)[5:12])
	var reply []string
	args := struct{}{}
	err := client.Call("Peer.GetPeers", args, &reply)
	if err != nil {
		log.Fatal("Peers error:", err)
	}
	p.Peers.lock.Lock()
	for _, addr := range reply {
		if addr != p.Address && p.Peers.Clients[addr] == nil {
			p.connectToPeer(addr)
		}
	}
	p.Peers.lock.Unlock()
}

func (p *Peer) FloodJoinMessage() {
	for addr := range p.Peers.Clients {
		senderAddr := p.Address
		if addr == senderAddr {
			continue
		}

		if p.Peers.Clients[addr] == nil {
			p.connectToPeer(addr)
		}
		client := p.Peers.Clients[addr]
		var reply bool
		method := "Peer.JoinMessage"
		success := false
		for i := 0; i < 3; i++ {
			err := client.Call(method, senderAddr, &reply)
			if err == nil && reply {
				success = true
				break
			}
			log.Println("Failed to send Join Message message, retrying:", err)
		}
		if !success {
			log.Printf("Failed to send Join Message to %s after retries", addr)
		}
	}
}
