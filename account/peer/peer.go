package peer

import (
	L "account/ledger"
	"bufio"
	"fmt"
	"net"
)

type Peer struct {
	ID      string
	Ledger  *L.Ledger
	Peers   []string
	Adress  string
	Send    net.Conn
	Receive net.Conn
}

func (p *Peer) Connect(addr string) {
	fmt.Printf("I am peer %s\n", p.ID)
	conn, err := net.Dial("tcp", addr)
	if err != nil && conn == nil {
		p.receive()
	} else if err == nil {
		defer conn.Close()
		fmt.Printf("my address is %s\n", p.Adress)
		p.Send = conn
		fmt.Printf("Peer %s connected to %s\n", p.ID, addr)
		p.RequestPeers(conn)

		// scanner := bufio.NewScanner(conn)
		// for scanner.Scan() {
		// 	message := scanner.Text()
		// }
		p.receive()
	}
}

func (p *Peer) RequestPeers(connection net.Conn) {
	//request peers from a peer
	fmt.Fprintf(connection, "RequestPeers")
}

func (p *Peer) AddPeer(addr string) {
	p.Peers = append(p.Peers, addr)
}

func (p *Peer) receive() {
	fmt.Printf("Open for connections as peer %s\n", p.ID)
	ln, _ := net.Listen("tcp", p.Adress)
	fmt.Printf("my address is %s\n", ln.Addr().String())
	p.Adress = ln.Addr().String()
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		fmt.Printf("A peer from addr %s has connected to %s\n", conn.RemoteAddr().String(), p.ID)
		p.Receive = conn
		go p.handleConnection(conn)
	}
}

func (p *Peer) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Listening for messages %s\n", conn.LocalAddr().String())
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		if message == "RequestPeers" {
			fmt.Printf("Peer %s requested peers\n", p.ID)
			//send peers
			// for _, peer := range p.Peers {
			// 	fmt.Fprintf(conn, peer)
			// }
			continue
		}
		fmt.Printf("I (peer %s) got message: %s \n", p.ID, message)
		fmt.Fprintf(conn, "Hello from peer %s\n", p.ID)
	}
}

// func (p *Peer) FloodTransaction(tx *T.Transaction) {
// 	message, err := json.Marshal(tx)
// 	if err != nil {
// 		fmt.Println("Error marshaling transaction:", err)
// 		return
// 	}
// 	p.FloodMessage(string(message))
// }

func (p *Peer) FloodMessage(message string) {
	//broadcast message to all peers
	for _, peer := range p.Peers {
		if peer.Send != nil {
			fmt.Fprint(peer.Send, message+"\n")
		}
	}
}
