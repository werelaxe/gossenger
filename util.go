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
