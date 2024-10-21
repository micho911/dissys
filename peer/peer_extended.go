package peer

import (
	"crypto/sha256"
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
	hashedMsg := sha256.Sum256([]byte(msg))
	hashedMsgSlice := hashedMsg[:]

	decodedKey, err := encrypt.DecodeKey(t.From)
	util.Must(err)

	validSignature := encrypt.Verify(hashedMsgSlice, []byte(t.Signature), decodedKey)

	if validSignature {
		l.Accounts[t.From] -= t.Amount
		l.Accounts[t.To] += t.Amount
	}
}
