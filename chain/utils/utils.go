package utils

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

func GenMethod(func_name string) string {
	hash := sha3.NewLegacyKeccak256()
	s := []byte(func_name)
	hash.Write(s)
	buf := hash.Sum(nil)
	return hex.EncodeToString(buf)
}
