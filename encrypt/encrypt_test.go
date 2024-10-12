package encrypt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRSAandAESTest(t *testing.T) {
	t.Log("Generating RSA key pair")
	pk, sk := KeyGen(128)
	msg, _ := rand.Prime(rand.Reader, 64)
	if pk.N.Cmp(&sk.N) != 0 {
		t.Fatal("RSA public and private keys do not have the same modulus")
	}
	if pk.N.BitLen() != 128 {
		t.Fatal("RSA public key modulus is not 128 bits")
	}

	t.Logf("Original RSA message: %s", msg.String())

	cipher := Encrypt(msg, pk)
	t.Logf("RSA Ciphertext: %s", cipher.String())

	decMsg := Decrypt(&cipher, sk)
	t.Logf("RSA Decrypted message: %s", decMsg.String())

	if decMsg.Cmp(msg) != 0 {
		t.Fatal("RSA decrypted message does not match the original message")
	}

	t.Log("Running AES test with RSA private key parts")

	aeskey, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")

	plaintext := []byte(sk.Exp.String() + "," + sk.N.String())
	fileName := "test.txt"

	t.Log("Encrypting RSA secret key to file")
	EncryptToFile(fileName, plaintext, aeskey)

	t.Log("Decrypting from file")
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

	t.Log("Decrypting RSA message using the decrypted RSA secret key")
	rsaDecMsg := Decrypt(&cipher, Key{d, n})

	if rsaDecMsg.Cmp(msg) != 0 {
		t.Fatal("Decrypted RSA message does not match the original message after AES encryption/decryption")
	}

	t.Logf("Test successful: decrypted RSA message: %s", rsaDecMsg.String())

	os.Remove("test.txt")
	os.Remove("test_decrypted.txt")
}

func TestSignAndVerify(t *testing.T) {
	// Generate RSA key pair (512 bits for simplicity)
	pk, sk := KeyGen(512)

	message := []byte("This is a test message for signing and verification")
	t.Log("Original message:", message)

	signature := Sign(message, sk)
	t.Log("RSA Signature:", signature)

	if !Verify(message, signature, pk) {
		t.Fatal("RSA signature verification failed")
	}
	t.Log("RSA signature verification passed for the correct message")

	modifiedMessage := []byte("This is a modified test message for signing and verification")
	t.Log("Modified message:", modifiedMessage)

	if Verify(modifiedMessage, signature, pk) {
		t.Fatal("RSA signature verification incorrectly passed for the modified message")
	}
	t.Log("RSA signature verification correctly failed for the modified message")
}

func TestSpeedHash(t *testing.T) {
	messageSize := 10 * 1024 //

	message := make([]byte, messageSize)
	_, err := rand.Read(message)
	if err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	startTime := time.Now()

	iterations := 10000
	for i := 0; i < iterations; i++ {
		_ = sha256.Sum256(message)
	}

	duration := time.Since(startTime)
	totalBitsHashed := uint64(messageSize * 8 * iterations)
	bitsPerSecond := float64(totalBitsHashed) / duration.Seconds()

	t.Logf("Hashed %d bits in %v\n", totalBitsHashed, duration)
	t.Logf("Hashing speed: %.2f bits per second\n", bitsPerSecond)
}

func TestSpeedRSA(t *testing.T) {
	// Measure the time to generate the 2000-bit RSA key pair
	startKeyGen := time.Now()
	_, sk := KeyGen(2000)
	keyGenDuration := time.Since(startKeyGen)

	t.Logf("RSA 2000-bit key generation time: %v", keyGenDuration)

	// Define the message and hash it
	message := []byte("This is a test message for measuring RSA signature speed")
	hash := sha256.Sum256(message)
	hashedMessage := hash[:]

	// Measure the time to sign the hash
	startSign := time.Now()
	signature := Sign(hashedMessage, sk)
	signDuration := time.Since(startSign)

	// Calculate the total number of bits processed (SHA-256 produces a 256-bit hash)
	totalBitsProcessed := 256

	// Calculate the speed in bits per second
	bitsPerSecond := float64(totalBitsProcessed) / signDuration.Seconds()

	t.Logf("RSA signature produced in: %v", signDuration)
	t.Logf("RSA Signature: %x", signature)
	t.Logf("RSA signing speed: %.2f bits per second", bitsPerSecond)
}

func TestSpeedRSAWithNoHash(t *testing.T) {
	// Measure the time to generate the 2000-bit RSA key pair
	startKeyGen := time.Now()
	_, sk := KeyGen(2000)
	keyGenDuration := time.Since(startKeyGen)

	t.Logf("RSA 2000-bit key generation time: %v", keyGenDuration)

	// Define a large message (e.g., 10KB)
	messageSize := 10 * 1024 // 10KB in bytes
	message := make([]byte, messageSize)
	_, err := rand.Read(message)
	if err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	// Calculate how many 2000-bit chunks we need to process
	// 2000 bits = 250 bytes per RSA operation
	chunkSize := 2000 / 8
	numChunks := len(message) / chunkSize
	if len(message)%chunkSize != 0 {
		numChunks++
	}

	// Measure the time to sign the entire message using RSA
	startRSA := time.Now()
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end > len(message) {
			end = len(message)
		}

		// Extract each chunk of the message
		chunk := new(big.Int).SetBytes(message[start:end])

		// Sign the chunk (use Decrypt with sk as signing)
		_ = Decrypt(chunk, sk)
	}
	signDuration := time.Since(startRSA)

	// Calculate speed in bits per second
	totalBitsProcessed := uint64(len(message) * 8)
	bitsPerSecond := float64(totalBitsProcessed) / signDuration.Seconds()

	t.Logf("RSA signing time for full message: %v", signDuration)
	t.Logf("Processed %d bits using RSA for signing", totalBitsProcessed)
	t.Logf("RSA signing speed: %.2f bits per second", bitsPerSecond)
}
