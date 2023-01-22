package internal

import (
	"math/rand"
	"strings"
	"time"
)

var randomWords = []string{
	"humpty",
	"dumpty",
	"horsey",
	"donkey",
	"yikes",
	"monkey",
	"doo",
	"scooby",
	"dooby",
	"vanilla",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomWord() string {
	return randomWords[rand.Intn(len(randomWords))]
}

func randomString() string {
	return strings.Join(
		[]string{
			randomWord(),
			randomWord(),
			randomWord(),
			randomWord(),
			randomWord(),
			randomWord(),
		},
		" ",
	)
}

func randomStringShort() string {
	return randomWord()
}

func randomStringsShort(n int) []string {
	return shuffle(randomWords, n)
}

func shuffle(vals []string, n int) []string {
	if n > len(vals) {
		panic("don't have so many words")
	}

	ret := make([]string, n)

	for i, randIndex := range rand.Perm(len(vals))[:n] {
		ret[i] = vals[randIndex]
	}

	return ret
}

func randomStrings(n int) []string {
	l := make([]string, n)

	for i := range l {
		l[i] = randomString()
	}

	return l
}

func randomWordRand(rnd *rand.Rand) string {
	return randomWords[rnd.Intn(len(randomWords))]
}

func randomStringRand(rnd *rand.Rand) string {
	return randomWordRand(rnd)
}

func randomLongStringRand(rnd *rand.Rand) string {
	l := make([]string, 6)

	for i := range l {
		l[i] = randomWordRand(rnd)
	}

	return strings.Join(l, "-")
}

func randomStringsRand(n int, rnd *rand.Rand) []string {
	return shuffleRand(randomWords, n, rnd)
}

func randomLongStringsRand(n int, rnd *rand.Rand) []string {
	l := make([]string, n)

	for i := range l {
		l[i] = randomLongStringRand(rnd)
	}

	return l
}

func shuffleRand(vals []string, n int, rnd *rand.Rand) []string {
	if n > len(vals) {
		panic("don't have so many words")
	}

	ret := make([]string, n)

	for i, randIndex := range rnd.Perm(len(vals))[:n] {
		ret[i] = vals[randIndex]
	}

	return ret
}
