package encrypt

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"
)

func TestRSAandAESTest(t *testing.T) {
	log.Println("Generating RSA key pair")
	pk, sk := KeyGen(128)
	msg, _ := rand.Prime(rand.Reader, 64)
	log.Printf("Original RSA message: %s", msg.String())

	cipher := Encrypt(msg, pk)
	log.Printf("RSA Ciphertext: %s", cipher.String())

	decMsg := Decrypt(&cipher, sk)
	log.Printf("RSA Decrypted message: %s", decMsg.String())

	if decMsg.Cmp(msg) != 0 {
		t.Fatal("RSA decrypted message does not match the original message")
	}

	log.Println("Running AES test with RSA private key parts")

	aeskey, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")

	plaintext := []byte(sk.exp.String() + "," + sk.n.String())
	fileName := "test.txt"

	log.Println("Encrypting RSA secret key to file")
	EncryptToFile(fileName, plaintext, aeskey)

	log.Println("Decrypting from file")
	DecryptFromFile(fileName, aeskey)

	decryptedContent, err := os.ReadFile("test_decrypted.txt")
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	decryptedString := string(decryptedContent)
	parts := strings.Split(decryptedString, ",")
	if len(parts) != 2 {
		t.Fatal("Decrypted content does not have the expected format (exp,n)")
	}

	var d, n big.Int
	if _, success := d.SetString(parts[0], 10); !success {
		t.Fatal("Failed to convert exp to big.Int")
	}
	if _, success := n.SetString(parts[1], 10); !success {
		t.Fatal("Failed to convert n to big.Int")
	}

	log.Println("Decrypting RSA message using the decrypted RSA secret key")
	rsaDecMsg := Decrypt(&cipher, Key{d, n})

	if rsaDecMsg.Cmp(msg) != 0 {
		t.Fatal("Decrypted RSA message does not match the original message after AES encryption/decryption")
	}

	log.Printf("Test successful: decrypted RSA message: %s", rsaDecMsg.String())

	os.Remove("test.txt")
	os.Remove("test_decrypted.txt")
}
