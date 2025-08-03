package mqtt

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

const (
	uniqueIDLength = 6

	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func generateClientId(appID string) string {
	b := make([]byte, uniqueIDLength)
	var src = rand.NewSource(time.Now().UnixNano())

	// A src.Int63() generates 63 random bits, enough for letterIdxMax
	// characters!
	for i, cache, remain := uniqueIDLength-1, src.Int63(),
		letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	// #nosec G103
	return fmt.Sprintf("%s-%s", appID, *(*string)(unsafe.Pointer(&b)))
}
