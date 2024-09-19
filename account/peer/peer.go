package peer

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Peer struct {
	Id      string
	Address string
	Peers   map[string]*rpc.Client
}

func (p *Peer) Connect(addr string) {
	p.Peers = make(map[string]*rpc.Client)
	client, err := rpc.Dial("tcp", addr)
	if err != nil || client == nil {
		p.serve()
	} else {
		defer client.Close()
		fmt.Printf("Peer %s connected to %s\n", p.Address, addr)
		p.Peers[addr] = client
		p.serve()
		p.requestPeers(client)
		p.FloodMessage("JoinMessage")
	}

}

func (p *Peer) serve() {
	rpc.Register(p)

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
			go rpc.ServeConn(conn)
		}
	}()
}

/* func (p *Peer) FloodMessage(msg string) {
	for addr := range p.Peers {
		if p.Peers[addr] == nil {
			continue
		}
		fmt.Printf("Peer %s sending to client %p\n", p.Id, p.Peers[addr])
		var reply bool
		senderAddr := p.Address
		method := fmt.Sprintf("Peer.%s", msg)
		err := p.Peers[addr].Call(method, senderAddr, &reply)
		if err != nil || !reply {
			log.Fatal("FloodMessage error: ", msg, err)
		}
	}
} */

func (p *Peer) FloodMessage(msg string) {
	for addr, client := range p.Peers {
		if client == nil {
			continue
		}
		fmt.Printf("Peer %s sending to client %p at %s\n", p.Id, client, addr)
		var reply bool
		senderAddr := p.Address
		method := fmt.Sprintf("Peer.%s", msg)

		// Retry logic for sending messages
		for i := 0; i < 3; i++ {
			err := client.Call(method, senderAddr, &reply)
			if err == nil && reply {
				break // Success
			}
			log.Println("Failed to send message, retrying:", err)
		}
	}
}

func (p *Peer) requestPeers(client *rpc.Client) {
	var reply []string
	args := struct{}{}
	fmt.Printf("Peer %s requesting from client %p\n", p.Id, client)
	err := client.Call("Peer.GetPeers", args, &reply)
	if err != nil {
		log.Fatal("Peers error:", err)
	}
	for _, addr := range reply {
		if addr != p.Address && p.Peers[addr] == nil {
			client, err := rpc.Dial("tcp", addr)
			if err != nil {
				log.Println("Could not connect to requested peer")
			} else {
				p.Peers[addr] = client
				fmt.Printf("Added address %s to peer %s map\n", addr, p.Id)
			}
		}
	}
	fmt.Printf("Peer %s new map: %+v\n", p.Id, p.Peers)
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

func (p *Peer) ConnectionCount() int {
	count := 0
	for _, client := range p.Peers {
		if client != nil {
			count++
		}
	}
	return count
}
