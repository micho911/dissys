package ledger

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encrypt"
	"fmt"
	"strconv"
	"sync"
)

type AU = int

type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
}

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount AU
}

type SignedTransaction struct {
	ID        string
	From      string
	To        string
	Amount    AU
	Signature string
	Time      int64
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) UndoTransaction(tx *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Accounts[tx.From] += tx.Amount
	l.Accounts[tx.To] -= tx.Amount
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

	if tx.From == "BlockChain" { //Could make a vk sk pair for the genesis block
		l.Accounts[tx.From] -= tx.Amount
		l.Accounts[tx.To] += tx.Amount
		return
	}

	msg := tx.ID + tx.From + tx.To + strconv.Itoa(tx.Amount) + strconv.FormatInt(tx.Time, 10)

	decodedKey, err := encrypt.DecodeKey(tx.From)
	if err != nil {
		fmt.Println("Error decoding public key:", err)
		return
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(tx.Signature)
	if err != nil {
		fmt.Println("Error decoding signature:", err)
		fmt.Println(tx.Signature)
		return
	}

	validSignature := encrypt.Verify([]byte(msg), signatureBytes, decodedKey)

	if validSignature {
		newBalanceSender := l.Accounts[tx.From] - tx.Amount
		newBalanceReceiver := l.Accounts[tx.To] + tx.Amount

		if newBalanceSender < 0 || newBalanceReceiver < 0 {
			return
		}
		l.Accounts[tx.From] = newBalanceSender
		l.Accounts[tx.To] = newBalanceReceiver
	}
}

func SerializeTransactions(transactions []*SignedTransaction) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(transactions); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func DeserializeTransactions(data []byte) ([]*SignedTransaction, error) {
	var transactions []*SignedTransaction
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	if err := decoder.Decode(&transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}
