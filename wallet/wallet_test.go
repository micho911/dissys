package wallet

import (
	"encoding/json"
	"encrypt"
	"math/big"
	"testing"
	"time"
)

func TestWallet(t *testing.T) {
	filename := "wallet_test_file"
	password := "strongpassword"
	message := []byte("This is a test message.")

	pubKeyStr := Generate(filename, password)
	t.Log("Public Key:", pubKeyStr)

	// Sign the message
	signature := Sign(filename, password, message)
	t.Log("Signature:", signature)

	var pkData map[string]string
	err := json.Unmarshal([]byte(pubKeyStr), &pkData)
	if err != nil {
		t.Fatal(err)
	}

	var e, n big.Int
	e.SetString(pkData["e"], 10)
	n.SetString(pkData["n"], 10)
	pk := encrypt.Key{Exp: e, N: n}

	valid := encrypt.Verify(message, signature, pk)
	if !valid {
		t.Fatal("Signature verification failed")
	} else {
		t.Log("Signature verification succeeded")
	}

}

func TestIncorrectPasswordTiming(t *testing.T) {
	filename := "wallet_invalid_test_file"
	correctPassword := "strongpassword"
	incorrectPasswords := []string{"wrongpass1", "wrongpass2", "wrongpass3", "wrongpass4", "wrongpass5"}
	message := []byte("This is a test message.")

	pubKeyStr := Generate(filename, correctPassword)
	t.Log("Public Key:", pubKeyStr)

	startTime := time.Now()
	signature := Sign(filename, correctPassword, message)
	correctDuration := time.Since(startTime)
	t.Logf("Time taken for correct password attempt: %v", correctDuration)

	var pkData map[string]string
	err := json.Unmarshal([]byte(pubKeyStr), &pkData)
	if err != nil {
		t.Fatal(err)
	}

	var e, n big.Int
	e.SetString(pkData["e"], 10)
	n.SetString(pkData["n"], 10)
	pk := encrypt.Key{Exp: e, N: n}

	valid := encrypt.Verify(message, signature, pk)
	if !valid {
		t.Fatal("Signature verification failed")
	} else {
		t.Log("Signature verification succeeded")
	}

	for i, pwd := range incorrectPasswords {
		startTime = time.Now()
		func(pwd string, i int) {
			defer func() {
				if r := recover(); r != nil {
					duration := time.Since(startTime)
					t.Logf("Attempt %d with incorrect password took: %v", i+1, duration)
				}
			}()
			_ = Sign(filename, pwd, message)
			t.Fatalf("Attempt %d: Expected panic due to incorrect password, but Sign succeeded", i+1)
		}(pwd, i)
	}
}
