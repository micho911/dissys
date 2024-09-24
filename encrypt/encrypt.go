package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
)

type Key struct {
	exp big.Int
	n   big.Int
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

		pk.exp = *e
		pk.n = n
		sk.exp = d
		sk.n = n
	}
	fmt.Println("n is", n.String())
	fmt.Println("the length of n is", n.BitLen())
	fmt.Println("d is", d.String())
	fmt.Println("the length of d is", d.BitLen())
	return pk, sk
}

func Encrypt(m *big.Int, pk Key) big.Int {
	var temp big.Int
	return *temp.Exp(m, &pk.exp, &pk.n)
}

func Decrypt(c *big.Int, sk Key) big.Int {
	var temp big.Int
	return *temp.Exp(c, &sk.exp, &sk.n)
}

func EncryptToFile(fileToWrite string, message []byte, key []byte) {
	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(c)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, message, nil)

	combined := append(nonce, ciphertext...)

	err = os.WriteFile(fileToWrite, combined, 0644)
	if err != nil {
		panic(err.Error())
	}

}

func DecryptFromFile(fileToRead string, key []byte) {
	fileContent, err := os.ReadFile(fileToRead)
	if err != nil {
		panic(err.Error())
	}

	if len(fileContent) < 12 {
		panic("File content is too short to contain a valid nonce and ciphertext")
	}
	nonce := fileContent[:12]
	ciphertext := fileContent[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	// write the decrypted content to test_decrypted.txt
	err = os.WriteFile("test_decrypted.txt", plaintext, 0644)
	if err != nil {
		panic(err.Error())
	}
}
