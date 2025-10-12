package util

import (
	"math/rand"
)

const CharSet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const CharSetLength = 26 + 26 + 10

var randomStringMap = map[int]string{}

func GenerateRandomString(length int) string {
	if _, ok := randomStringMap[length]; ok {
		return randomStringMap[length]
	}
	rs := ""
	for i := 0; i < length; i++ {
		idx := rand.Intn(CharSetLength)
		rs = rs + string(CharSet[idx])
	}
	randomStringMap[length] = rs
	return randomStringMap[length]
}
