package peer

import (
	L "account/ledger"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

type Peer struct {
	ID      string
	Ledger  *L.Ledger
	Peers   *PeerSet
	Address string
	Conns   *rpc.Client
}

type PeerSet map[string]struct{}

type JoinMessage struct {
	PeerID string
	Peers  *PeerSet
}

func (p *Peer) Connect(addr string) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil && client == nil {
		p.startNetwork()
	} else if err == nil {
		defer client.Close()
		p.Conns = client
		p.serve()
		time.Sleep(1 * time.Second)
		p.syncPeers()
	} else {
		log.Fatal("Error connecting:", err)
	}
}

func (p *Peer) SendPeersToNewPeer(args struct{}, peers *PeerSet) error {
	*peers = *p.Peers
	return nil
}

func (p *Peer) syncPeers() {
	incomingPeers := NewPeerSet()
	err := p.Conns.Call("Peer.SendPeersToNewPeer", struct{}{}, incomingPeers)
	time.Sleep(1 * time.Second)
	if err != nil {
		log.Fatal("Error calling SendPeersToNewPeer:", err)
	}
	p.Peers = incomingPeers
	p.Peers.addPeer(p.Address)
	p.FloodMessage(JoinMessage{PeerID: p.ID, Peers: (p.Peers)})
}

func (p *Peer) startNetwork() {
	p.Peers = NewPeerSet()
	p.serve()
	p.Peers.addPeer(p.Address)
}

func (p *Peer) serve() {

	gob.Register(JoinMessage{})
	rpc.Register(p)

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	p.Address = ln.Addr().String()
	fmt.Println("Peer address:", p.Address)
	go func() {
		for {
			conn, err := ln.Accept()
			fmt.Println("Connection accepted")
			if err != nil {
				log.Println("Connection error:", err)
				continue
			}
			go rpc.ServeConn(conn) // Serve RPC calls using the connection
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

func (p *Peer) FloodMessage(message interface{}) {
	switch msg := message.(type) {
	case JoinMessage:
		err := p.Conns.Call("Peer.ReceiveJoinMessage", msg, nil)
		time.Sleep(1 * time.Second)
		if err != nil {
			log.Println("Error flooding JoinMessage:", err)
		}
	// case TransactionMessage:
	// 		fmt.Println("Flooding TransactionMessage")
	// 		// Similar logic for TransactionMessage
	default:
		fmt.Println("Unknown message type")
	}
}

func (p *Peer) ReceiveJoinMessage(joinMsg JoinMessage, reply *bool) error {
	for _, peer := range joinMsg.Peers.Members() {
		p.Peers.addPeer(peer)
	}
	*reply = true
	return nil
}

func NewPeerSet() *PeerSet {
	peers := make(PeerSet) // Initialize the map
	return &peers
}

func (peers *PeerSet) addPeer(peerAddr string) {
	(*peers)[peerAddr] = struct{}{}
}

func (peers *PeerSet) Members() []string {
	fmt.Printf("membering peers")
	members := make([]string, 0, len(*peers))
	for peer := range *peers {
		members = append(members, peer)
	}
	return members
}
