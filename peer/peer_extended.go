package peer

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

func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

func (l *Ledger) SignedTransaction(t *SignedTransaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	msg := t.ID + t.From + t.To + strconv.Itoa(t.Amount)

	decodedKey, err := encrypt.DecodeKey(t.From)
	if err != nil {
		fmt.Println("Error decoding public key:", err)
		return
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(t.Signature)
	if err != nil {
		fmt.Println("Error decoding signature:", err)
		return
	}

	validSignature := encrypt.Verify([]byte(msg), signatureBytes, decodedKey)

	if validSignature {
		l.Accounts[t.From] -= t.Amount
		l.Accounts[t.To] += t.Amount
	} else {
		fmt.Println("Invalid signature for transaction ID:", t.ID)
	}
}
