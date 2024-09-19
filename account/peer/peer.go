package peer

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

type Peer struct {
	Id      string
	Address string
	Peers   map[string]*rpc.Client
	Ledger  *Ledger
}

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

func (p *Peer) Connect(addr string) {
	p.Peers = make(map[string]*rpc.Client)
	client, err := rpc.Dial("tcp", addr)
	if err != nil || client == nil {
		p.serve()
	} else {
		p.Peers[addr] = client
		fmt.Printf("Peer %s connected to %s\n", p.Id, addr)
		p.serve()
		p.requestPeers(client)
		p.FloodMessage("JoinMessage")
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
		// Retry logic for sending messages
		for i := 0; i < 3; i++ {
			err := client.Call(method, senderAddr, &reply)
			if err == nil && reply {
				success = true
				break // Success
			}
			log.Println("Failed to send message, retrying:", err)
		}
		if !success {
			log.Printf("Failed to send message %s to %s after retries", msg, addr)
		}
	}
}

func (p *Peer) connectToPeer(addr string) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil || client == nil {
		fmt.Printf("Peer %s could not connect to %s\n", p.Id, addr)
	}
	p.Peers[addr] = client
	fmt.Printf("Peer %s connected to %s\n", p.Id, addr)
}
func (p *Peer) FloodTransaction(tx *Transaction) {
	p.Ledger.Transaction(tx)
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
		method := "Peer.UpdateLedger"
		success := false
		// Retry logic for sending transactions
		for i := 0; i < 3; i++ {
			err := client.Call(method, tx, &reply)
			if err == nil && reply {
				success = true
				break // Success
			}
			log.Println("Failed to send transaction, retrying:", err)
		}
		if !success {
			log.Printf("Failed to send transaction to %s after retries", addr)
		}
	}
}

func (p *Peer) UpdateLedger(tx *Transaction, reply *bool) error {
	p.Ledger.Transaction(tx)
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
