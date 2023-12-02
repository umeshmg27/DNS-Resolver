package resolver

import (
	"math/rand"
	"time"
)

func generateRandomNumber() uint16 {
	rand.NewSource(time.Now().UnixNano())
	return uint16(rand.Intn(65536)) // 65536 is the exclusive upper bound for a 2-byte number
}
