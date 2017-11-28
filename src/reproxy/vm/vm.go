package vm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	uuid "github.com/satori/go.uuid"
)

type VirtualMachine struct {
	otto *otto.Otto
	used map[string]bool
}

func log(level string, s ...interface{}) {
	fmt.Printf("%s\t%s\n", level, fmt.Sprint(s...))
}

var lib = map[string]interface{}{
	"aes": map[string]interface{}{
		"encrypt": func(a, b, c string) interface{} {
			return "TODO THIS :)"
		},
		"decrypt": func(text, secret, iv string) interface{} {

			text_b := []byte(text)
			secret_b := []byte(secret)
			iv_b := []byte(iv)

			block, err := aes.NewCipher(secret_b)
			if err != nil {
				return otto.NullValue()
			}

			if len(text_b) < aes.BlockSize {
				// Crypted too sort
				return otto.NullValue()
			}

			// CBC mode always works in whole blocks.
			if len(text_b)%aes.BlockSize != 0 {
				// crypted is not a multiple of the block size
				return otto.NullValue()
			}

			mode := cipher.NewCBCDecrypter(block, iv_b)

			// CryptBlocks can work in-place if the two arguments are the same.
			decoded := make([]byte, len(text_b))
			mode.CryptBlocks(decoded, text_b)

			var unpad = func(src []byte) []byte {
				length := len(src)
				unpadding := int(src[length-1])
				return src[:(length - unpadding)]
			}

			decoded = unpad(decoded)

			return string(decoded)
		},
	},
	"base64": map[string]interface{}{
		"encode": func(s string) interface{} {
			src := []byte(s)

			return base64.URLEncoding.EncodeToString(src)
		},
		"decode": func(s string) interface{} {
			text, err := base64.StdEncoding.DecodeString(s)
			if nil != err {
				return otto.NullValue()
			}

			return string(text)
		},
	},
	"hash": map[string]interface{}{
		"md5": func(s interface{}) interface{} {
			text := s.(string)
			bytes := []byte(text)

			md5_adder := md5.New()
			md5_adder.Write(bytes)

			return hex.EncodeToString(md5_adder.Sum(nil))
		},
		"sha1": func(s interface{}) interface{} {
			text := s.(string)
			bytes := []byte(text)

			sha1_adder := sha1.New()
			sha1_adder.Write(bytes)

			return hex.EncodeToString(sha1_adder.Sum(nil))
		},
		"hmacSha1": func(s interface{}, k interface{}) interface{} {
			text := s.(string)
			key := k.(string)
			bytes := []byte(text)
			key_for_sign := []byte(key)

			h := hmac.New(sha1.New, key_for_sign)
			h.Write(bytes)

			return hex.EncodeToString(h.Sum(nil))
		},
		"sha256": func(s interface{}) interface{} {
			text := s.(string)
			bytes := []byte(text)

			sha256_adder := sha256.New()
			sha256_adder.Write(bytes)

			return hex.EncodeToString(sha256_adder.Sum(nil))
		},
	},
	"log": func(s ...interface{}) {
		log("INFO", s...)
	},
	"http": func(args ...interface{}) interface{} {
		/*Arg list:
		1: HTTP method
		2: URL
		3: Headers map
		4: BODY
		5: Client private key
		6: Client cert
		7: Allow insecure certificates (true or false as string). Default = True
		8: Timeout
		*/
		l := len(args)

		if l < 2 {
			return nil
		}
		method := strings.ToUpper(args[0].(string))
		urlStr := args[1].(string)

		headers := map[string]interface{}{}
		if l >= 3 {
			headers = args[2].(map[string]interface{})
		}

		body_reader := strings.NewReader("")
		if l >= 4 {
			body_reader = strings.NewReader(args[3].(string))
		}

		//Support for client certificates
		client := http.DefaultClient
		if l >= 6 && args[4].(string) != "" && args[5].(string) != "" {
			keyBlock, _ := pem.Decode([]byte(args[4].(string)))
			if keyBlock == nil {
				//Bad privkey
				log("ERROR", "Bad private key")
				return "Bad private key"
			}
			keyPEM := pem.EncodeToMemory(keyBlock)

			certBlock, _ := pem.Decode([]byte(args[5].(string)))
			if certBlock == nil {
				//Bad cert
				log("ERROR", "Bad certificate")
				return "Bad certificate"
			}
			certPEM := pem.EncodeToMemory(certBlock)

			cert, err := tls.X509KeyPair(certPEM, keyPEM)
			if err != nil {
				log("ERROR", fmt.Sprint("Error X509KeyPair", err.Error()))
				return otto.NullValue()
			}

			// Setup HTTPS client
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}

			// Check if asked to enable insecure
			if l >= 7 {
				tlsConfig.InsecureSkipVerify, _ = args[6].(bool)
			}

			tlsConfig.BuildNameToCertificate()
			transport := &http.Transport{TLSClientConfig: tlsConfig}
			client = &http.Client{Transport: transport}
		}

		timeout := 10 * time.Second
		if len(args) > 7 && args[7].(int64) != 0 {
			timeout = time.Duration(args[7].(int64)) * time.Second
		}

		client.Timeout = time.Duration(timeout)

		req, _ := http.NewRequest(method, urlStr, body_reader)

		for k, v := range headers {
			req.Header.Add(k, v.(string))
		}

		res, err := client.Do(req)
		if nil != err {
			log("ERROR", "Client.Do error - No connectivity", err.Error())
			return otto.NullValue()
		}

		body, body_err := ioutil.ReadAll(res.Body)
		if nil != body_err {
			log("ERROR", "Error while trying to read response body", err.Error())
			return otto.NullValue()
		}

		response_headers := map[string]string{}

		for k, v := range res.Header {
			response_headers[k] = v[0]
		}

		return map[string]interface{}{
			"status":     res.StatusCode,
			"statusText": res.Status,
			"headers":    response_headers,
			"body":       string(body),
		}
	},
	"rsa": map[string]interface{}{
		"encrypt": func(a, b, c string) interface{} {
			return "TODO THIS :)"
		},
		"decryptPKCS1v15": func(b64msg, privkeypem string) interface{} {

			block, _ := pem.Decode([]byte(privkeypem))
			if block == nil {
				//bad key data: not PEM-encoded
				log("ERROR", "bad key data: not PEM-encoded")
				return otto.NullValue()
			}

			privkey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				//Bad privkey
				log("ERROR", "bad privkey")
				return otto.NullValue()
			}

			rsaprivkey := privkey.(*rsa.PrivateKey)
			message, err := base64.StdEncoding.DecodeString(b64msg)
			if err != nil {
				//Bad message
				log("ERROR", "bad message")
				return otto.NullValue()
			}

			cleartext, err := rsa.DecryptPKCS1v15(rand.Reader, rsaprivkey, message)
			if err != nil {
				//Bad crypt
				log("ERROR", "bad crypt")
				return otto.NullValue()
			}

			return string(cleartext)
		},
	},
	"uuid": map[string]interface{}{
		"version1": func() string {
			return "not implemented"
		},
		"version2": func() string {
			return "not implemented"
		},
		"version3": func() string {
			return "not implemented"
		},
		"version4": func() string {
			return uuid.NewV4().String()
		},
		"version5": func() string {
			return "not implemented"
		},
	},
	"sleep": map[string]interface{}{
		"seconds": func(s interface{}) bool {

			if n, err := interface2number(s); nil == err {
				time.Sleep(time.Duration(n*1000) * time.Millisecond)
				return true
			}

			return false
		},
		"milliseconds": func(s interface{}) bool {

			if n, err := interface2number(s); nil == err {
				time.Sleep(time.Duration(n) * time.Millisecond)
				return true
			}

			return false
		},
	},
}

func New() *VirtualMachine {

	v := &VirtualMachine{
		otto: otto.New(),
		used: map[string]bool{},
	}

	v.initialize()

	return v
}

func (v *VirtualMachine) initialize() {

	v.otto.Set("use", func(s interface{}) {
		name := s.(string)
		v.Use(name)
	})

	v.Use("log")
}

func (v *VirtualMachine) Run(src string) (otto.Value, error) {
	return v.otto.Run(src)
}

func (v *VirtualMachine) Use(name string) {

	if imported, exists := v.used[name]; exists && imported {
		return
	}

	if src, exist := lib[name]; exist {
		v.used[name] = true
		v.otto.Set(name, src)
	}

}

func (v *VirtualMachine) Set(name string, value interface{}) error {
	return v.otto.Set(name, value)
}
