package registrar

import (
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type randGenerator struct {
	rg *rand.Rand
}

func newRandGenerator() *randGenerator {
	toRet := &randGenerator{rg: rand.New(rand.NewSource(time.Now().UnixNano()))}
	return toRet
}

func (r *randGenerator) RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.rg.Intn(len(letterRunes))]
	}
	return string(b)
}
