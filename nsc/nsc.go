package nsc

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encrypt"
	"fmt"
	"ledger"
	"math/big"
	"strconv"
	"sync"
	"time"
)

var hardness = getHardness()

const slotDuration = int64(1)

const seed = 100

const blockSize = 10

var genesisTime = int64(1733007600)

var genesisVk, genesisSk = encrypt.KeyGen(2048)

var genesisWallets []Wallet

type Wallet struct {
	PublicKey  encrypt.Key
	PrivateKey encrypt.Key
}

type Block struct {
	Transactions []*ledger.SignedTransaction
	Draw         Draw
	Ptr          [32]byte
	Sig          []byte
}

type Node struct {
	Block    *Block
	Parent   *Node
	Children []*Node
}

type Tree struct {
	Seed      int
	Root      *Node
	BlockSize int
	Lock      sync.Mutex
	Orphans   []*Block
}

type Draw struct {
	Vk        encrypt.Key
	TimeStamp int64
	Slot      int64
	DrawSign  []byte
}

func getHardness() *big.Int {
	var twoTo279 = new(big.Int).Lsh(big.NewInt(1), 275)
	var twoTo280 = new(big.Int).Lsh(big.NewInt(1), 276)
	var difference = new(big.Int).Sub(twoTo280, twoTo279)
	var halfway = new(big.Int).Div(difference, big.NewInt(7))
	return new(big.Int).Add(twoTo279, halfway)
}

func GenerateGenesisWallets() {
	for i := 0; i < 10; i++ {
		vk, sk := encrypt.KeyGen(2048)
		genesisWallets = append(genesisWallets, Wallet{PublicKey: vk, PrivateKey: sk})
	}
}

func GetGenesisWallets() *[]Wallet {
	return &genesisWallets
}

func GenesisBlock() *Block {
	transactions := []*ledger.SignedTransaction{}
	for _, wallet := range genesisWallets {
		tx := &ledger.SignedTransaction{
			ID:     "genesis",
			From:   "BlockChain",
			To:     encrypt.EncodeKey(wallet.PublicKey),
			Amount: 1000000,
			Time:   genesisTime,
		}
		msg := tx.ID + tx.From + tx.To + strconv.Itoa(tx.Amount) + strconv.FormatInt(tx.Time, 10)
		sig := encrypt.Sign([]byte(msg), genesisSk)

		tx.Signature = base64.StdEncoding.EncodeToString(sig)

		transactions = append(transactions, tx)
	}

	draw := &Draw{
		TimeStamp: genesisTime,
		Slot:      0,
		Vk:        genesisVk,
	}

	str := "lottery"
	slotStr := strconv.FormatInt(draw.Slot, 10)
	seed := seed
	seedStr := strconv.Itoa(seed)

	drawStr := str + slotStr + seedStr
	drawSign := encrypt.Sign([]byte(drawStr), genesisSk)

	draw.DrawSign = drawSign

	return &Block{
		Transactions: transactions,
		Draw:         *draw,
		Ptr:          [32]byte{},
		Sig:          []byte("Root"),
	}
}

func makeRoot() *Node {
	return &Node{Block: GenesisBlock(), Parent: nil, Children: []*Node{}}
}

func NewTree() *Tree {
	return &Tree{Seed: seed, Root: makeRoot(), BlockSize: blockSize, Lock: sync.Mutex{}}
}

func (t *Tree) AddBlock(block *Block) bool {
	parent := t.GetParentNode(t.Root, block.Ptr)
	if parent == nil {
		t.Orphans = append(t.Orphans, block)
		return false
	}

	newNode := &Node{Block: block, Parent: parent, Children: []*Node{}}
	parent.Children = append(parent.Children, newNode)

	t.AddOrphans(newNode)

	return true
}

func SerializeBlock(block *Block) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(block); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	if err := decoder.Decode(&block); err != nil {
		return nil, err
	}

	return &block, nil
}

func (t *Tree) AddOrphans(parent *Node) {
	var updatedOrphans []*Block

	for _, orphanBlock := range t.Orphans {
		serializedParentBlock, _ := SerializeBlock(parent.Block)
		parentBlockHash := sha256.Sum256(serializedParentBlock)

		if parentBlockHash == orphanBlock.Ptr {
			fmt.Print("An orphan block was inserted into the tree\n")
			orphanNode := &Node{Block: orphanBlock, Parent: parent, Children: []*Node{}}
			parent.Children = append(parent.Children, orphanNode)
		} else {
			updatedOrphans = append(updatedOrphans, orphanBlock)
		}
	}

	t.Orphans = updatedOrphans
}

func (t *Tree) GetParentNode(node *Node, ptr [32]byte) *Node {
	serializedBlock, _ := SerializeBlock(node.Block)
	blockHash := sha256.Sum256(serializedBlock)

	if blockHash == ptr {
		return node
	}

	var parent *Node
	for _, child := range node.Children {
		parent = t.GetParentNode(child, ptr)
		if parent != nil {
			return parent
		}
	}
	return parent
}

func (t *Tree) FindLeadingNode() *Node {
	return findLeadingNode(t.Root, 0).Node
}

type result struct {
	Node  *Node
	Depth int
}

func findLeadingNode(node *Node, depth int) result {
	if node == nil {
		return result{nil, depth}
	}

	if len(node.Children) == 0 {
		return result{node, depth}
	}

	leadingResult := result{node, depth}

	for _, child := range node.Children {
		childResult := findLeadingNode(child, depth+1)

		if childResult.Depth > leadingResult.Depth {
			leadingResult = childResult
		} else if childResult.Depth == leadingResult.Depth {
			if bytes.Compare(childResult.Node.Block.Sig, leadingResult.Node.Block.Sig) < 0 {
				leadingResult = childResult
			}
		}
	}

	return leadingResult
}

func MakeDraw(vk encrypt.Key, timeStamp int64, slot int64, drawSign []byte) Draw {
	return Draw{
		Vk:        vk,
		TimeStamp: timeStamp,
		Slot:      slot,
		DrawSign:  drawSign,
	}
}

func (t *Tree) getSeed() int {
	return t.Seed
}

func (d Draw) GetSlot() int64 {
	return d.Slot
}

func (d Draw) getMsg(seed int) string {
	str := "lottery"
	slotStr := strconv.FormatInt(d.Slot, 10)
	seedStr := strconv.Itoa(seed)
	return str + slotStr + seedStr

}

func (t *Tree) GetCurrentSlot() (int64, int64) {
	currentTime := time.Now().Unix() // Current Unix timestamp in seconds
	elapsed := currentTime - t.GetGenesisTime()
	return elapsed / slotDuration, currentTime
}

func (t *Tree) GetGenesisTime() int64 {
	return t.Root.Block.Draw.TimeStamp
}

func (d Draw) getVal(t *Tree) big.Int {
	hash := sha256.Sum256(d.DrawSign)
	num := new(big.Int)
	num.SetBytes(hash[:])
	tickets := t.getTickets(d.Vk)
	return *num.Mul(num, big.NewInt(int64(tickets)))
}

func (t *Tree) getTickets(vk encrypt.Key) int {
	initialTx := t.Root.Block.Transactions

	for _, tx := range initialTx {
		strVk := encrypt.EncodeKey(vk)
		if tx.To == strVk {
			return tx.Amount
		}
	}
	return 0
}

func (d *Draw) IsWinner(tree *Tree) bool {
	drawVal := d.getVal(tree)

	return drawVal.Cmp(hardness) > 0
}

func (d *Draw) verify(tree *Tree) bool {
	if !d.IsWinner(tree) {
		return false
	}

	seed := tree.getSeed()

	drawMsg := d.getMsg(seed)
	return encrypt.Verify([]byte(drawMsg), d.DrawSign, d.Vk)
}

func (b *Block) Verify(tree *Tree) bool {
	if !b.Draw.verify(tree) {
		fmt.Printf("Block draw could not be verified\n")
		return false
	}

	for _, tx := range b.Transactions {
		msg := tx.ID + tx.From + tx.To + strconv.Itoa(tx.Amount) + strconv.FormatInt(tx.Time, 10)
		decodedKey, _ := encrypt.DecodeKey(tx.From)
		signatureBytes, _ := base64.StdEncoding.DecodeString(tx.Signature)
		if !encrypt.Verify([]byte(msg), signatureBytes, decodedKey) {
			fmt.Printf("A block transaction could not be verified\n")
			return false
		}
	}

	msg, _ := ledger.SerializeTransactions(b.Transactions)
	return encrypt.Verify(msg, b.Sig, b.Draw.Vk)
}
