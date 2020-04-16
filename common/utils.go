package common

import (
	"crypto/md5"
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

func Hash(str string) []byte {
	hash := md5.Sum([]byte(str))
	return hash[:]
}

func InitRandom() {
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

func Unique(slice []uint) map[uint]bool {
	keys := make(map[uint]bool)
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
		}
	}
	return keys
}

func GetLimitAndOffset(query url.Values) (int, int, error) {
	var limit, offset int
	var err error

	rawLimit, ok := query["limit"]
	if !ok {
		limit = DefaultApiLimit
	} else {
		limit, err = strconv.Atoi(rawLimit[0])
		if err != nil {
			return 0, 0, errors.New("can not get limit and offset: " + err.Error())
		}
	}

	rawOffset, ok := query["limit"]
	if !ok {
		offset = 0
	} else {
		offset, err = strconv.Atoi(rawOffset[0])
		if err != nil {
			return 0, 0, errors.New("can not get limit and offset: " + err.Error())
		}
	}
	return limit, offset, nil
}
