package wallet

import (
	"crypto/rand"
	"encoding/json"
	"encrypt"
	"math/big"
	"os"
	"util"

	"golang.org/x/crypto/scrypt"
)

type EncryptedKeyFile struct {
	Salt          []byte
	EncryptedData []byte
}

func Generate(filename string, password string) encrypt.Key {
	pk, sk := encrypt.KeyGen(2048)

	skData := map[string]string{
		"d": sk.Exp.String(),
		"n": sk.N.String(),
	}
	skBytes, err := json.Marshal(skData)
	util.Must(err)

	salt := make([]byte, 16)
	_, err = rand.Read(salt)
	util.Must(err)

	key, err := scrypt.Key([]byte(password), salt, 1<<20, 8, 1, 32)
	util.Must(err)

	err = os.WriteFile(filename, salt, 0644)
	util.Must(err)

	encrypt.EncryptToFile(filename, skBytes, key, true)

	return pk
}

func Sign(filename string, password string, msg []byte) []byte {
	fileContent, err := os.ReadFile(filename)
	util.Must(err)

	if len(fileContent) < 16+12 {
		panic("File content is too short to contain valid data")
	}
	salt := fileContent[:16]
	encryptedData := fileContent[16:]

	key, err := scrypt.Key([]byte(password), salt, 1<<20, 8, 1, 32)
	util.Must(err)

	skBytes := encrypt.DecryptFromData(encryptedData, key)

	var skData map[string]string
	err = json.Unmarshal(skBytes, &skData)
	util.Must(err)

	var d, n big.Int
	d.SetString(skData["d"], 10)
	n.SetString(skData["n"], 10)

	sk := encrypt.Key{Exp: d, N: n}

	signature := encrypt.Sign(msg, sk)

	return signature
}
