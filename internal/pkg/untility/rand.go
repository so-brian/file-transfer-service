package untility

import (
	"math/rand"
	"time"
)

func RandStr(length int) string {
	rand.New(rand.NewSource(time.Now().Unix()))

	ran_str := make([]byte, length)

	// Generating Random string
	for i := 0; i < length; i++ {
		ran_str[i] = byte(rune(65 + rand.Intn(25)))
	}

	// Displaying the random string
	str := string(ran_str)

	return str
}
