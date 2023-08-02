package navicat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/blowfish"
	"golang.org/x/sys/windows/registry"
	"strings"
)

type Version string

const HighVersionKey = "libcckeylibcckey"
const HighVersionIv = "libcciv libcciv "
const LowVersionHexKey = "42ceb271a5e458b74aea93947922354391873340"
const LowVersionHexIv = "d9c7c3c8870d64bd"

const NsPath = `Software\PremiumSoft\Navicat\Servers`

type Server struct {
	Path                string
	ServerVersion       uint64
	Host                string
	Port                uint64
	UserName            string
	HighVersionPassword string
	LowVersionPassword  string
	SshHost             string
	SshPort             uint64
	SshUserName         string
	SshPassword         string
	NavicatVersion      string
}

func newHighVersionCipher() (Cipher, error) {
	return &HighVersionCipher{
		key: []byte(HighVersionKey),
		iv:  []byte(HighVersionIv),
	}, nil
}

func newLowVersionCipher() (Cipher, error) {
	key, _ := hex.DecodeString(LowVersionHexKey)
	iv, _ := hex.DecodeString(LowVersionHexIv)
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &LowVersionCipher{
		key:    key,
		iv:     iv,
		cipher: block,
	}, nil
}

func GetNavicatServers() ([]*Server, error) {
	var hnc Cipher
	var lnc Cipher
	hnc, _ = newHighVersionCipher()
	lnc, _ = newLowVersionCipher()
	nsp, err := registry.OpenKey(registry.CURRENT_USER, NsPath, registry.READ)
	defer nsp.Close()
	var exists bool
	if nil != err {
		exists = false
	} else {
		exists = true
	}
	var servers []*Server
	if exists {
		subKeys, err := nsp.ReadSubKeyNames(999)
		if err != nil && err.Error() != "EOF" {
			return nil, err
		}
		for _, subKey := range subKeys {
			server, success := handleServerConf(subKey, hnc, lnc)
			if success {
				servers = append(servers, server)
			}
		}
	}
	return servers, nil
}

func handleServerConf(subKey string, hnc Cipher, lnc Cipher) (*Server, bool) {
	serverPath := strings.Join([]string{NsPath, subKey}, `\`)
	sp, err := registry.OpenKey(registry.CURRENT_USER, serverPath, registry.READ)
	if err != nil {
		fmt.Printf("handle path %s error: %v\n", serverPath, err)
		return nil, false
	}
	defer sp.Close()
	v, _, _ := sp.GetIntegerValue("ServerVersion")
	h, _, _ := sp.GetStringValue("Host")
	un, _, _ := sp.GetStringValue("UserName")
	pwd, _, _ := sp.GetStringValue("Pwd")
	port, _, _ := sp.GetIntegerValue("Port")
	sshh, _, _ := sp.GetStringValue("SSH_Host")
	sshun, _, _ := sp.GetStringValue("SSH_UserName")
	sshpwd, _, _ := sp.GetStringValue("SSH_Password")
	sshport, _, _ := sp.GetIntegerValue("SSH_Port")
	if len(h) > 0 {
		hpw := pwd
		lpw := pwd
		if len(pwd) > 0 {
			defer func() {
				if r := recover(); r != nil {
					// ignore
				}
			}()
			hpw, _ = hnc.Decrypt(pwd)
			lpw, _ = lnc.Decrypt(pwd)
		}
		return &Server{
			Path:                subKey,
			ServerVersion:       v,
			Host:                h,
			UserName:            un,
			HighVersionPassword: hpw,
			LowVersionPassword:  lpw,
			Port:                port,
			SshHost:             sshh,
			SshUserName:         sshun,
			SshPassword:         sshpwd,
			SshPort:             sshport,
		}, true
	}
	return nil, false
}

type Cipher interface {
	Encrypt(content string) (string, error)

	Decrypt(content string) (string, error)
}

type NoneCipher struct {
}

func (n *NoneCipher) Encrypt(input string) (string, error) {
	return input, nil
}

func (n *NoneCipher) Decrypt(input string) (string, error) {
	return input, nil
}

type HighVersionCipher struct {
	key []byte
	iv  []byte
}

func (h *HighVersionCipher) Encrypt(input string) (string, error) {
	realInput := []byte(input)
	aseCipher, err := aes.NewCipher(h.key)
	if err != nil {
		return "", err
	}
	content := PKCS5Padding(realInput, aseCipher.BlockSize())
	result := make([]byte, len(content))
	aseEncrypter := cipher.NewCBCEncrypter(aseCipher, h.iv)
	aseEncrypter.CryptBlocks(result, content)
	return strings.ToUpper(hex.EncodeToString(result)), nil
}

func (h *HighVersionCipher) Decrypt(input string) (string, error) {
	realInput, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}
	result := make([]byte, len(realInput))
	aseCipher, err := aes.NewCipher(h.key)
	if err != nil {
		return "", err
	}
	aesDecrypter := cipher.NewCBCDecrypter(aseCipher, h.iv)
	aesDecrypter.CryptBlocks(result, realInput)
	unPadding, err := PKCS5UnPadding(result)
	if err != nil {
		return "", err
	}
	return string(unPadding), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func PKCS5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	unPadding := int(origData[length-1])

	if length < unPadding {
		return nil, fmt.Errorf("invalid unpadding length")
	}
	return origData[:(length - unPadding)], nil
}

type LowVersionCipher struct {
	key    []byte
	cipher cipher.Block
	iv     []byte
}

func (l *LowVersionCipher) Encrypt(input string) (string, error) {
	return "", errors.New("unsupported encrypt method")
}

func (l *LowVersionCipher) Decrypt(input string) (string, error) {
	ciphertext, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}
	if len(ciphertext)%8 != 0 {
		return "", errors.New("ciphertext length must be a multiple of 8")
	}
	plaintext := make([]byte, len(ciphertext))
	cv := make([]byte, len(l.iv))
	copy(cv, l.iv)
	blocksLen := len(ciphertext) / blowfish.BlockSize
	leftLen := len(ciphertext) % blowfish.BlockSize
	decrypter := NewECBDecrypter(l.cipher)
	for i := 0; i < blocksLen; i++ {
		temp := make([]byte, blowfish.BlockSize)
		copy(temp, ciphertext[i*blowfish.BlockSize:(i+1)*blowfish.BlockSize])
		if err != nil {
			panic(err)
		}
		decrypter.CryptBlocks(temp, temp)
		xorBytes(temp, cv)
		copy(plaintext[i*blowfish.BlockSize:(i+1)*blowfish.BlockSize], temp)
		for j := 0; j < len(cv); j++ {
			cv[j] ^= ciphertext[i*blowfish.BlockSize+j]
		}
	}

	if leftLen != 0 {
		decrypter.CryptBlocks(cv, cv)
		temp := make([]byte, leftLen)
		copy(temp, ciphertext[blocksLen*blowfish.BlockSize:])
		xorBytes(temp, cv[:leftLen])
		copy(plaintext[blocksLen*blowfish.BlockSize:], temp)
	}

	return string(plaintext), nil
}

func xorBytes(a []byte, b []byte) {
	for i := 0; i < len(a); i++ {
		aVal := int(a[i]) & 0xff // convert byte to integer
		bVal := int(b[i]) & 0xff
		a[i] = byte(aVal ^ bVal) // xor aVal and bVal and typecast to byte
	}
}
