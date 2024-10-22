package ledger

import (
	"encrypt"
	"strconv"
	"testing"
)

func TestValidSignedTransaction(t *testing.T) {
	pk1, sk1 := encrypt.KeyGen(2048)
	pk2, _ := encrypt.KeyGen(2048)

	enc_pk1, enc_pk2 := encrypt.EncodeKey(pk1), encrypt.EncodeKey(pk2)

	ledger1 := MakeLedger()
	ledger2 := MakeLedger()

	//Create ledgers for pk1 and pk2
	ledger1.Accounts[enc_pk1], ledger1.Accounts[enc_pk2] = 50, 80
	ledger2.Accounts[enc_pk1], ledger2.Accounts[enc_pk2] = 30, 60

	//Create valid signed transaction from pk1 to pk2
	msg := "1" + enc_pk1 + enc_pk2 + strconv.Itoa(15)
	valid_sig := encrypt.Sign([]byte(msg), sk1)
	valid_tx := SignedTransaction{
		ID:        "1",
		From:      enc_pk1,
		To:        enc_pk2,
		Amount:    15,
		Signature: string(valid_sig),
	}

	//Perform transaction on ledger1 and ledger2
	ledger1.SignedTransaction(&valid_tx)
	ledger2.SignedTransaction(&valid_tx)

	if ledger1.Accounts[enc_pk1] != 35 {
		t.Errorf("Ledger1 has %d in pk1's account, expected 35", ledger1.Accounts[enc_pk1])
	}

	if ledger2.Accounts[enc_pk1] != 15 {
		t.Errorf("Ledger2 has %d in pk1's account, expected 15", ledger2.Accounts[enc_pk1])
	}

	if ledger1.Accounts[enc_pk2] != 95 {
		t.Errorf("Ledger1 has %d in pk2's account, expected 95", ledger1.Accounts[enc_pk2])
	}

	if ledger2.Accounts[enc_pk2] != 75 {
		t.Errorf("Ledger2 has %d in pk2's account, expected 75", ledger2.Accounts[enc_pk2])
	}
}

func TestInvalidSignedTransaction(t *testing.T) {
	pk1, _ := encrypt.KeyGen(2048)   // Key pair for pk1
	pk2, sk2 := encrypt.KeyGen(2048) // Key pair for pk2 (we'll use sk2 to simulate an invalid signature)

	enc_pk1, enc_pk2 := encrypt.EncodeKey(pk1), encrypt.EncodeKey(pk2)

	ledger1 := MakeLedger()
	ledger2 := MakeLedger()

	// Create ledgers for pk1 and pk2
	ledger1.Accounts[enc_pk1], ledger1.Accounts[enc_pk2] = 50, 80
	ledger2.Accounts[enc_pk1], ledger2.Accounts[enc_pk2] = 30, 60

	// Create invalid signed transaction from pk1 to pk2 (signature is generated using the wrong key sk2)
	msg := "1" + enc_pk1 + enc_pk2 + strconv.Itoa(15)
	invalid_sig := encrypt.Sign([]byte(msg), sk2) // Using sk2 instead of sk1 for signing
	invalid_tx := SignedTransaction{
		ID:        "1",
		From:      enc_pk1,
		To:        enc_pk2,
		Amount:    15,
		Signature: string(invalid_sig),
	}

	ledger1.SignedTransaction(&invalid_tx)
	ledger2.SignedTransaction(&invalid_tx)

	// Check that the account balances have not changed
	if ledger1.Accounts[enc_pk1] != 50 {
		t.Errorf("Ledger1 has %d in pk1's account, expected 50 (no change due to invalid transaction)", ledger1.Accounts[enc_pk1])
	}

	if ledger2.Accounts[enc_pk1] != 30 {
		t.Errorf("Ledger2 has %d in pk1's account, expected 30 (no change due to invalid transaction)", ledger2.Accounts[enc_pk1])
	}

	if ledger1.Accounts[enc_pk2] != 80 {
		t.Errorf("Ledger1 has %d in pk2's account, expected 80 (no change due to invalid transaction)", ledger1.Accounts[enc_pk2])
	}

	if ledger2.Accounts[enc_pk2] != 60 {
		t.Errorf("Ledger2 has %d in pk2's account, expected 60 (no change due to invalid transaction)", ledger2.Accounts[enc_pk2])
	}
}
