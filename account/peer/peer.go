package peer

import (
	L "account/ledger"
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Peer struct {
	ID      string
	Ledger  *L.Ledger
	Peers   PeerList
	Adress  string
	Send    net.Conn
	Receive net.Conn
}

type PeerList struct {
	Peers []string
}

func (p *Peer) Connect(addr string) {
	fmt.Printf("I am peer %s\n", p.ID)
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil && client == nil {
		p.server()
	} else if err == nil {
		defer client.Close()
		fmt.Printf("my address is %s\n", p.Adress)
		fmt.Printf("Peer %s connected to %s\n", p.ID, addr)
		var peers PeerList
		err = client.Call("Peer.SendPeersToNewPeer", p, &peers)
		if err != nil {
			log.Fatal("Error calling SendPeersToNewPeer:", err)
		}
		p.Adress = addr
		p.Peers = peers
		p.Peers.addPeer(p.Adress)
		p.Peers = p.Peers.removeDuplicatePeers()
		p.server()
	} else {
		log.Fatal("Error connecting:", err)
	}
}

func (p *Peer) SendPeersToNewPeer(newPeer *Peer, peers *PeerList) error {
	peers = &p.Peers
	peers.addPeer(p.Adress)
	peers.removeDuplicatePeers()
	return nil
}

func (p *Peer) server() {
	fmt.Printf("Open for connections as peer %s\n", p.ID)
	rpc.Register(p)

	ln, err := net.Listen("tcp", p.Adress)
	if err != nil {
		log.Fatal("Error listening:", err)
	}

	p.Adress = ln.Addr().String()
	p.Peers.addPeer(p.Adress)
	fmt.Printf("my address is %s\n", p.Adress)

	// Manually serve RPC connections instead of using `rpc.HandleHTTP()`
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Connection error:", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
}

// func (p *Peer) FloodTransaction(tx *T.Transaction) {
// 	message, err := json.Marshal(tx)
// 	if err != nil {
// 		fmt.Println("Error marshaling transaction:", err)
// 		return
// 	}
// 	p.FloodMessage(string(message))
// }

// func (p *Peer) FloodMessage(message string) {
// 	//broadcast message to all peers
// 	for _, peer := range p.Peers {
// 		if peer.Send != nil {
// 			fmt.Fprint(peer.Send, message+"\n")
// 		}
// 	}
// }

func (peers *PeerList) addPeer(peer string) {
	peers.Peers = append(peers.Peers, peer)
}

func (peers *PeerList) removeDuplicatePeers() PeerList {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range peers.Peers {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return PeerList{list}
}
