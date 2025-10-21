package util

import (
	"math/rand"
	"sync"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	mu   sync.RWMutex
	pool = make(map[int]string)
)

func GenerateRandomString(n int) string {
	mu.RLock()
	if s, ok := pool[n]; ok {
		mu.RUnlock()
		return s
	}
	mu.RUnlock()

	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}
	s := string(b)

	mu.Lock()
	pool[n] = s
	mu.Unlock()
	return s
}
