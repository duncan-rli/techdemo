package main

import (
	"io"
	"log"
	"net/http"
	"fmt"
	"io/ioutil"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"encoding/json"
)


// Commands to create certs
//
// openssl genrsa -out server.key 2048
// openssl ecparam -genkey -name secp384r1 -out server.key
// openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
//
//

var store = make(map[string][]byte)

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize + len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[0:32])
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)

	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func StoreServer(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	jData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("ReadAll: ", err)
		return
	}

	res := map[string][]byte{}
	json.Unmarshal(jData, &res)
	id := res["id"]
	data := res["payload"]

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		s := `{"error":"Key Gen Error: ` + err.Error() + `"}`
		io.WriteString(w, s)
		return
	}

	fmt.Println("key ", key)

	var encData []byte
	encData, err = encrypt(key, data)
	if err != nil {
		s := `{"error":"Encryption Error: ` + err.Error() + `"}`
		io.WriteString(w, s)
		return
	}


	// write to store
	store[string(id)] = encData
	skey := base64.StdEncoding.EncodeToString(key)

	s := `{"aesKey":"` + skey + `"}`

	// reply to client
	io.WriteString(w, s)
}

func RetrieveServer(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	jData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("ReadAll: ", err)
		return
	}

	res := map[string][]byte{}
	json.Unmarshal(jData, &res)
	fmt.Println("res ", res)
	id := res["id"]
	key := res["aesKey"]

	var decKey []byte

	decKey, err = base64.StdEncoding.DecodeString(string(key))

	// data from store
	if store[string(id)] == nil {
		s := `{"error":"ID not found in store"}`
		io.WriteString(w, s)
		return
	}

	entry := store[string(id)]

	encData := make([]byte, len(entry))
	copy(encData, entry)

	var decData []byte
	decData, err = decrypt(decKey, encData)

	if err != nil {
		s := `{"error":"Decryption Error: ` + err.Error() + `"}`
		io.WriteString(w, s)
		return
	}

	// reply to client with data
	skey := base64.StdEncoding.EncodeToString(decData)
	s := `{"payload":"` + skey + `"}`
	fmt.Println("s", s)
	io.WriteString(w, s)
}

func main() {
	http.HandleFunc("/store", StoreServer)
	http.HandleFunc("/retrieve", RetrieveServer)
	err := http.ListenAndServeTLS(":8443", "/etc/myapp/ssl/tmp/server.pem", "/etc/myapp/ssl/tmp/server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
