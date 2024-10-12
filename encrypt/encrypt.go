package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"math/big"
	"os"
	"util"
)

type Key struct {
	Exp big.Int
	N   big.Int
}

func KeyGen(k int) (pk Key, sk Key) {
	var d, n, zero big.Int
	for d.Cmp(&zero) == 0 {
		p, _ := rand.Prime(rand.Reader, k/2)
		q, _ := rand.Prime(rand.Reader, k/2)
		e := new(big.Int).SetInt64(3)

		n.Mul(p, q)

		var pMinusOne big.Int
		var qMinusOne big.Int
		pMinusOne.Sub(p, new(big.Int).SetInt64(1))
		qMinusOne.Sub(q, new(big.Int).SetInt64(1))

		var qTimesP big.Int
		qTimesP.Mul(&pMinusOne, &qMinusOne)
		d.ModInverse(e, &qTimesP)

		pk.Exp = *e
		pk.N = n
		sk.Exp = d
		sk.N = n
	}
	return pk, sk
}

func Encrypt(m *big.Int, pk Key) big.Int {
	var temp big.Int
	return *temp.Exp(m, &pk.Exp, &pk.N)
}

func Decrypt(c *big.Int, sk Key) big.Int {
	var temp big.Int
	return *temp.Exp(c, &sk.Exp, &sk.N)
}

func EncryptToFile(fileToWrite string, message []byte, key []byte, options ...bool) {
	c, err := aes.NewCipher(key)
	util.Must(err)

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(c)
	util.Must(err)

	ciphertext := aesgcm.Seal(nil, nonce, message, nil)
	combined := append(nonce, ciphertext...)

	appendMode := false
	if len(options) > 0 {
		appendMode = options[0]
	}
	var file *os.File
	if appendMode {
		file, err = os.OpenFile(fileToWrite, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	} else {
		file, err = os.Create(fileToWrite)
	}
	util.Must(err)
	defer file.Close()
	_, err = file.Write(combined)
	util.Must(err)

}

func DecryptFromData(encryptedData []byte, key []byte) []byte {
	if len(encryptedData) < 12 {
		panic("Encrypted data is too short")
	}
	nonce := encryptedData[:12]
	ciphertext := encryptedData[12:]

	block, err := aes.NewCipher(key)
	util.Must(err)

	aesgcm, err := cipher.NewGCM(block)
	util.Must(err)

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	util.Must(err)

	return plaintext
}

func DecryptFromFile(fileToRead string, key []byte) {
	fileContent, err := os.ReadFile(fileToRead)
	util.Must(err)

	plaintext := DecryptFromData(fileContent, key)

	// write the decrypted content to test_decrypted.txt
	err = os.WriteFile("test_decrypted.txt", plaintext, 0644)
	util.Must(err)
}

func Sign(message []byte, sk Key) []byte {
	hash := sha256.Sum256(message)
	m := new(big.Int).SetBytes(hash[:])
	if m.Cmp(&sk.N) >= 0 {
		panic("Hash too large to sign with the provided RSA key")
	}
	signature := Decrypt(m, sk)

	return signature.Bytes()
}

func Verify(message []byte, signature []byte, pk Key) bool {
	hash := sha256.Sum256(message)

	signatureInt := new(big.Int).SetBytes(signature)
	recoveredHashInt := Encrypt(signatureInt, pk)
	recoveredHash := recoveredHashInt.Bytes()
	originalHash := hash[:]

	return string(recoveredHash) == string(originalHash)
}
