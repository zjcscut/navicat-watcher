package navicat

import (
	"fmt"
	"testing"
)

func TestGetNavicatServers(t *testing.T) {
	cipher := &HighVersionCipher{
		key: []byte(HighVersionKey),
		iv:  []byte(HighVersionIv),
	}
	r, _ := cipher.Decrypt("503AA930968F877F04770B47DD731DC0")
	fmt.Println(r)
	r, _ = cipher.Encrypt("root")
	fmt.Println(r)
	r, _ = cipher.Decrypt(r)
	fmt.Println(r)
}
