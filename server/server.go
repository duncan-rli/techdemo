package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	if len(key) < 32 {
		return nil, errors.New("Key too short")
	}
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
	var errorStr string

	jData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorStr = "ReadAll: " + err.Error()
	}

	res := map[string][]byte{}

	// create map to unmarshal into
	if len(errorStr) == 0 {
		err = json.Unmarshal(jData, &res)
		if err != nil {
			errorStr = "Json UnMarshal error"
		}
	}
	// get data from unmarshal result
	id := res["id"]
	data := res["data"]

	key := make([]byte, 32)
	if len(errorStr) == 0 {
		_, err = rand.Read(key)
		if err != nil {
			errorStr = `"Key Gen Error: ` + err.Error() + `"`
		}
	}
	// encrypt the data
	var encData []byte
	if len(errorStr) == 0 {
		encData, err = encrypt(key, data)
		if err != nil {
			errorStr = `Encryption Error: ` + err.Error() + `"`
		}
	}

	if len(errorStr) == 0 {
		// hash the id
		sum256Id := sha256.Sum256([]byte(id))
		sum256IdStr := base64.StdEncoding.EncodeToString([]byte(sum256Id[:]))

		// write to store
		store[sum256IdStr] = encData
	}

	// form response message
	respData := map[string][]byte{"aesKey": key, "error": []byte(errorStr)}
	jRespData, err := json.Marshal(respData)
	if err != nil {
		fmt.Println("Json Marshal failed")
		return
	}

	// reply to client
	jRespDataStr := string(jRespData)
	_, err = io.WriteString(w, jRespDataStr)
	if err != nil {
		fmt.Print("IO write error")
	}
	return
}

func RetrieveServer(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	jData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("ReadAll: ", err)
		return
	}

	var (
		errorStr, sum256IdStr string
	)

	res := map[string][]byte{}
	err = json.Unmarshal(jData, &res)
	if err != nil {
		errorStr = `"Json Unmarshall Error: ` + err.Error() + `"`
	}

	var (
		id  []byte
		key []byte
	)

	if len(errorStr) == 0 {
		id = res["id"]
		key = res["aesKey"]
	}

	// hash the id
	if len(errorStr) == 0 {
		sum256Id := sha256.Sum256([]byte(id))
		sum256IdStr = base64.StdEncoding.EncodeToString([]byte(sum256Id[:]))

		// check for entry in store
		if store[sum256IdStr] == nil {
			errorStr = `"ID not found in store"`
		}
	}

	var decData []byte

	// retrieve from store
	if len(errorStr) == 0 {
		entry := store[sum256IdStr]
		encData := make([]byte, len(entry))
		copy(encData, entry)

		// decrypt retrieved data
		decData, err = decrypt(key, encData)
		if err != nil {
			errorStr = `"Decryption Error: ` + err.Error() + `"`
		}
	}

	// form response message
	respData := map[string][]byte{"data": decData, "error": []byte(errorStr)}
	jRespData, err := json.Marshal(respData)
	if err != nil {
		fmt.Println("Json Marshal failed")
		return
	}

	// reply to client with data
	jRespDataStr := string(jRespData)
	_, err = io.WriteString(w, jRespDataStr)
	if err != nil {
		fmt.Print("IO write error")
	}
	return

}

func main() {
	http.HandleFunc("/store", StoreServer)
	http.HandleFunc("/retrieve", RetrieveServer)
	err := http.ListenAndServeTLS(":8443",
		"/etc/myapp/ssl/tmp/server.pem",
		"/etc/myapp/ssl/tmp/server.key",
		nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
