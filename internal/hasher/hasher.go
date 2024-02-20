package hasher

import (
	"crypto/md5"
	"encoding/base64"
	"strings"
)

func GetHashOfURL(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])
	hash := strings.Replace(encoded, "/", "", -1)[:8]
	return hash
}
