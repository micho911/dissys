package ledger

import (
	"encoding/base64"
	"encrypt"
	"fmt"
	"strconv"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
}

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

type SignedTransaction struct {
	ID        string
	From      string
	To        string
	Amount    int
	Signature string
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) Transaction(tx *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Accounts[tx.From] -= tx.Amount
	l.Accounts[tx.To] += tx.Amount
}

func (l *Ledger) SignedTransaction(tx *SignedTransaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	msg := tx.ID + tx.From + tx.To + strconv.Itoa(tx.Amount)

	decodedKey, err := encrypt.DecodeKey(tx.From)
	if err != nil {
		fmt.Println("Error decoding public key:", err)
		return
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(tx.Signature)
	if err != nil {
		fmt.Println("Error decoding signature:", err)
		return
	}

	validSignature := encrypt.Verify([]byte(msg), signatureBytes, decodedKey)

	if validSignature {
		l.Accounts[tx.From] -= tx.Amount
		l.Accounts[tx.To] += tx.Amount
	}
}
