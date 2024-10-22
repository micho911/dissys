package ledger

import (
	"encrypt"
	"strconv"
	"sync"
	"util"
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
	util.Must(err)

	validSignature := encrypt.Verify([]byte(msg), []byte(tx.Signature), decodedKey)

	if validSignature {
		l.Accounts[tx.From] -= tx.Amount
		l.Accounts[tx.To] += tx.Amount
	}
}
