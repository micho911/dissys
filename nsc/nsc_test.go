package nsc

import (
	"crypto/sha256"
	"ledger"
	"strconv"
	"testing"
)

func TestAddingBlockToTree(t *testing.T) {
	tree := NewTree()

	prevBlock := tree.FindLeadingNode().Block
	serializedPrevBlock, _ := SerializeBlock(prevBlock)
	prevBlockHash := sha256.Sum256(serializedPrevBlock)

	serializedRootBlock, _ := SerializeBlock(tree.Root.Block)
	rootHash := sha256.Sum256(serializedRootBlock)

	if rootHash != prevBlockHash {
		t.Fatalf("Incorrect ptr")
	}

	newBlock := &Block{
		Ptr: prevBlockHash,
	}

	parent := tree.GetParentNode(tree.Root, newBlock.Ptr)

	if parent == nil {
		t.Fatalf("Could not get parent of new block")
	}

	orphan := tree.AddBlock(newBlock)
	if !orphan {
		t.Fatalf("Did not add block to root")
	}

	n := len(tree.Root.Children)
	if n != 1 {
		t.Fatalf(`Children of root did not grow: %s`, strconv.Itoa(n))
	}

	if tree.Root.Children[0].Block != newBlock {
		t.Fatalf("New block is not in child of root")
	}
}

func TestAddingOrphanToTree(t *testing.T) {
	tree := NewTree()

	rootBlock := tree.FindLeadingNode().Block
	serializedRootBlock, _ := SerializeBlock(rootBlock)
	rootBlockHash := sha256.Sum256(serializedRootBlock)

	orphanParent := &Block{
		Ptr: rootBlockHash,
	}

	serializedOrphanParent, _ := SerializeBlock(orphanParent)
	parentHash := sha256.Sum256(serializedOrphanParent)

	orphan := &Block{
		Ptr: parentHash,
	}

	tree.AddBlock(orphan)

	if tree.Orphans[0] != orphan {
		t.Fatalf("Did not add orphan")
	}

	addedToTree := tree.AddBlock(orphanParent)
	if !addedToTree {
		t.Fatalf("Mistakenly registered orphan")
	}

	parentNode := tree.GetParentNode(tree.Root, orphan.Ptr)
	if parentNode == nil {
		t.Fatalf("Did not register parent for the orphan")
	}

	if len(tree.Orphans) != 0 {
		t.Fatalf("Did not attach orphan to tree")
	}
}

func TestRollBack(t *testing.T) {
	tree := NewTree()

	rootBlock := tree.FindLeadingNode().Block
	rootBlock.Transactions = []*ledger.SignedTransaction{
		{ID: "tx1", From: "Alice", To: "Bob", Amount: 10},
	}
	serializedRootBlock, _ := SerializeBlock(rootBlock)
	rootBlockHash := sha256.Sum256(serializedRootBlock)

	oldLeader := &Block{
		Ptr: rootBlockHash,
		Transactions: []*ledger.SignedTransaction{
			{ID: "tx2", From: "Charlie", To: "David", Amount: 20},
		},
	}

	addedToTree := tree.AddBlock(oldLeader)
	if !addedToTree {
		t.Fatalf("Did not add oldLeader to tree")
	}

	newLeaderParent := &Block{
		Ptr: rootBlockHash,
		Transactions: []*ledger.SignedTransaction{
			{ID: "tx3", From: "Eve", To: "Frank", Amount: 30},
		},
	}

	serializedNewLeaderParent, _ := SerializeBlock(newLeaderParent)

	newLeader := &Block{
		Ptr: sha256.Sum256(serializedNewLeaderParent),
		Transactions: []*ledger.SignedTransaction{
			{ID: "tx4", From: "George", To: "Hank", Amount: 40},
		},
	}

	addedToTree = tree.AddBlock(newLeader)
	if addedToTree {
		t.Fatalf("Did not add new leader to orphan set")
	}

	if oldLeader != tree.FindLeadingNode().Block {
		t.Fatalf("oldLeader is no longer the leading node")
	}

	addedToTree = tree.AddBlock(newLeaderParent)
	if !addedToTree {
		t.Fatalf("Could not add newLeaderParent to the tree")
	}

	if newLeader != tree.FindLeadingNode().Block {
		t.Fatalf("Did not perform rollback to newLeader")
	}
}
