package utils

import (
	"encoding/hex"
	"sync"

	"golang.org/x/crypto/sha3"
)

var wg sync.WaitGroup

func Go(f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}
func GoWait() {
	wg.Wait()
}

func GenMethod(func_name string) string {
	hash := sha3.NewLegacyKeccak256()
	s := []byte(func_name)
	hash.Write(s)
	buf := hash.Sum(nil)
	return hex.EncodeToString(buf)
}
