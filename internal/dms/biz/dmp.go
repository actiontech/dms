package biz

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"sync"
)

var buildinCaPrimaryKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDPr3VKlbvgP6NCRM4kuW+6wDEythhfxzVGgHG9Gthu3KUNxnd2
RiR8RwntEgsWum+dn1oYnsV5TdaA3Wg/mJrP7U3XFmX/OVl2UyqHMd+k/bFAJh8v
XKC2w9BtamzaSHdYro30skZ8jnxpjCD87+JtYgANySD0wRvlWwcD3slZ6wIDAQAB
AoGBAIPz/7q2rdrJtAm7u5n7s7BcwiVtKslXwVKc8ybqMo8lYz0AVxBvemj3nafh
aegz5gyonU69OcxblyjjA4Q8ikbhS6GOYxy27Oe6fYFjWzOWIMFkHe9QY7cqBkOL
8jzL0lzGyCWw+57l4h2tHPCctw/rfNi6NMFZUL2678H7u9rRAkEA/LlQ1WbZjHD2
dy6+xpPiut0dHrOG034uqHTv6P8NchT/eHjvy+5SDKiJc/nk1POZmwXWJXKcvFC6
zxwjvOAX2QJBANJgrbZRAIpwMJPNuOI+0ti3gIR7mpPZzwn0p2dJ4ORh48zp2b1M
QghibXGJMEHKcfU39H7v50/H/lZx0f0liWMCQQDR+ldDN/VBTwo49EnmTDFx+Q2c
2KUJTCoQJTjAakoNo4yv2CvFUPozMkUia1rJ5KyXtT28V4IKpTjRpBu9bqPhAkEA
nnIAAzMotBthCsDDQWrNhDlYiu9I8Zf2zem8dxd2UKvFVRy/SEn55bSz9vG7LaHa
iDS3aS8oSLc4wESDQiSWPwJADu2OQHUDDPE7hGU788Dsess2gY0xmJR6z36mWftD
Zz/GX75HZYICZBr6JjOVHHkLpByAWr5xonTLRyBhDqB7dg==
-----END RSA PRIVATE KEY-----`

var label = []byte("TVMvGp6rbgaBbWTU")

func getCaPrimaryKey() (string, error) {
	return buildinCaPrimaryKey, nil
}

var (
	decryptCache   = map[string]string{}
	decryptCacheMu = sync.RWMutex{}
)

func DecryptPassword(encrypted string) (string, error) {
	decryptCacheMu.RLock()
	if val, ok := decryptCache[encrypted]; ok {
		decryptCacheMu.RUnlock()
		return val, nil
	}
	decryptCacheMu.RUnlock()

	caKey, err := getCaPrimaryKey()
	if nil != err {
		return "", err
	}

	block, _ := pem.Decode([]byte(caKey))
	if nil == block {
		return "", errors.New("public GetKey error")
	}

	key, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	secretMessage, err := base64.StdEncoding.DecodeString(encrypted)
	if nil != err {
		return "", err
	}

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, secretMessage, label)
	if nil != err {
		return "", err
	}

	password := string(decrypted)

	decryptCacheMu.Lock()
	if 10240 < len(decryptCache) {
		decryptCache = map[string]string{encrypted: password}

	} else {
		decryptCache[encrypted] = password
	}
	decryptCacheMu.Unlock()

	return password, nil
}
