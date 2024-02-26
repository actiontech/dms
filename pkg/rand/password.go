package rand

import (
	"math/rand"
	"strings"
	"time"
)

func GenPassword(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(chars[r.Intn(len(chars))])
	}
	return b.String()
}


