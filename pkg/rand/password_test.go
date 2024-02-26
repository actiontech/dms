package rand

import (
	"math/rand"
	"testing"
)

func TestGenPassword(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n := rand.Intn(20) + 1 // n in the range of [1,20]
		p1 := GenPassword(n)

		if len(p1) != n {
			t.Errorf("Password length error: expected %d, actual %d", n, len(p1))
		}
		if len(p1) != n {
			t.Errorf("The length of the generated password is incorrect:%d != %d", len(p1), n)
		}
	}
}
