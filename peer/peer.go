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

type Peer struct {
	Id                string
	Address           string
	vk                encrypt.Key
	sk                encrypt.Key
	Peers             map[string]*rpc.Client
	Ledger            *ledger.Ledger
	Tree              *nsc.Tree
	transactionBuffer *TransactionBuffer
}

type TransactionBuffer struct {
	transactions []*ledger.SignedTransaction
	lock         sync.Mutex
}

func (p *Peer) Connect(addr string) {
	p.Peers = make(map[string]*rpc.Client)
	client, err := rpc.Dial("tcp", addr)
	if err != nil || client == nil {
		p.serve()
		p.joinLottery()
	} else {
		p.Peers[addr] = client
		fmt.Printf("Peer %s connected to %s\n", p.Id, addr)
		p.serve()
		p.requestPeers(client)
		p.FloodMessage("JoinMessage")
		p.joinLottery()
	}

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
	p.Peers[p.Address] = nil
	fmt.Printf("Peer %s serving at %s\n", p.Id, p.Address)

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
	go func() {
		previousSlot, _ := p.Tree.GetCurrentSlot()
		for {
			currentSlot, currentTime := p.Tree.GetCurrentSlot()

			if currentSlot > previousSlot {
				p.ProcessNewSlot(currentSlot, currentTime)
				previousSlot = currentSlot
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (p *Peer) MakeBlock(draw nsc.Draw) *nsc.Block {
	p.transactionBuffer.lock.Lock()
	defer p.transactionBuffer.lock.Unlock()

	insertionPoint := p.Tree.FindInsertionPoint(draw.GetSlot())

	transactions := p.transactionBuffer.transactions

	pendingTransactions := p.Tree.FilterPendingTransactions(insertionPoint, transactions)

	//remove transactions from buffer?

	serializedTransactions, _ := ledger.SerializeTransactions(pendingTransactions)
	sig := encrypt.Sign(serializedTransactions, p.sk)

	prevBlock := insertionPoint.Block
	serializedPrevBlock, _ := nsc.SerializeBlock(prevBlock)
	prevBlockHash := sha256.Sum256(serializedPrevBlock)

	newBlock := &nsc.Block{
		Transactions: pendingTransactions,
		Draw:         draw,
		Ptr:          prevBlockHash,
		Sig:          sig,
	}

	return newBlock
}

func (p *Peer) ProcessNewSlot(slot int64, time int64) {
	draw := p.draw(slot, time)
	if draw.IsWinner(p.Tree) {
		block := p.MakeBlock(draw)
		p.FloodBlock(block)
	}
}

func (p *Peer) FloodBlock(block *nsc.Block) {
	p.ProcessBlock(block, p.Address, nil)

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
		method := "Peer.ProcessBlock"
		args := []interface{}{block, p.Address}
		success := false
		for i := 0; i < 3; i++ {
			err := client.Call(method, args, &reply)
			if err == nil && reply {
				success = true
				break
			}
			log.Println("Failed to send block, retrying:", err)
		}
		if !success {
			log.Printf("Failed to send block to %s after retries", addr)
		}
	}
}

func (p *Peer) draw(slot int64, time int64) nsc.Draw {
	str := "lottery"
	slotStr := strconv.FormatInt(slot, 10)
	seed := p.Tree.Seed
	seedStr := strconv.Itoa(seed)
	drawStr := str + slotStr + seedStr
	drawSign := encrypt.Sign([]byte(drawStr), p.sk)
	draw := nsc.MakeDraw(p.vk, time, slot, drawSign)
	return draw
}

func (p *Peer) connectToPeer(addr string) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil || client == nil {
		fmt.Printf("Peer %s could not connect to %s\n", p.Id, addr)
	}
	p.Peers[addr] = client
}

func (p *Peer) requestPeers(client *rpc.Client) {
	var reply []string
	args := struct{}{}
	err := client.Call("Peer.GetPeers", args, &reply)
	if err != nil {
		log.Fatal("Peers error:", err)
	}
	for _, addr := range reply {
		if addr != p.Address && p.Peers[addr] == nil {
			p.connectToPeer(addr)
		}
	}
}

func (p *Peer) FloodMessage(msg string) {
	for addr := range p.Peers {
		senderAddr := p.Address
		if addr == senderAddr {
			continue
		}
		if p.Peers[addr] == nil {
			p.connectToPeer(addr)
		}
		client := p.Peers[addr]
		var reply bool
		method := fmt.Sprintf("Peer.%s", msg)
		success := false
		for i := 0; i < 3; i++ {
			err := client.Call(method, senderAddr, &reply)
			if err == nil && reply {
				success = true
				break
			}
			log.Println("Failed to send message, retrying:", err)
		}
		if !success {
			log.Printf("Failed to send message %s to %s after retries", msg, addr)
		}
	}
}

func (p *Peer) FloodTransaction(tx *ledger.SignedTransaction) {
	p.AddToTransactionBuffer(tx, p.Address, nil)

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
		method := "Peer.AddToTxBuffer"
		args := []interface{}{tx, p.Address}
		success := false
		for i := 0; i < 3; i++ {
			err := client.Call(method, args, &reply)
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
