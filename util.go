package main

import (
	"crypto/md5"
	"math/rand"
	"time"
)

func Hash(str string) []byte {
	hash := md5.Sum([]byte(str))
	return hash[:]
}

func initRandom() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func unique(slice []uint) map[uint]bool {
	keys := make(map[uint]bool)
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
		}
	}
	return keys
}
