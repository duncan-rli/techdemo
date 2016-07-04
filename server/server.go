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
	"os"
)


// Commands to create certs
//
// openssl genrsa -out server.key 2048
// openssl ecparam -genkey -name secp384r1 -out server.key                # WHAT DOES THIS LINE DO - IS IT NEEDED
// openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
//
//

var kkk []byte;
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
		fmt.Println("fail ciphertext")
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	fmt.Println("dec", key)
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

	fmt.Println("t",text)

	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		fmt.Println("decode data err")
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
fmt.Println("res ",res)
	id:= res["id"]
	data:= res["payload"]

	// data to secure store
fmt.Println("jd",string(jData))

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil{
		log.Println("Key Gen: ", err)
		return
	}

	fmt.Println("key ", key)

	var encData []byte
	encData, err = encrypt(key, data)
	if err != nil{
		log.Println("Encryption Error: ", err)
		return
	}

	// remove file if it already exists
	if _, err := os.Stat(string(id)); err == nil{
		_ = os.Remove(string(id))
	}
	// write to store
	store[string(id)] = encData
//	ioutil.WriteFile(string(id), encData, 0644)
	skey := base64.StdEncoding.EncodeToString(key)
	kkk = key
	s:=`{"aesKey":"`+skey+`"}`
	fmt.Println(s)

	// reply to client
	io.WriteString(w, s)
}

func RetrieveServer(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	jData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("ReadAll: ", err)
	}

	res := map[string][]byte{}
	json.Unmarshal(jData, &res)
	fmt.Println("res ",res)
	id:= res["id"]
//	var key = []byte {}
	key := res["aesKey"]

	var decKey []byte

	decKey, err = base64.StdEncoding.DecodeString(string(key))
//	var  i int
//	i=0
//	_, err = base64.StdEncoding.Decode(ds2, key)
//	if err != nil{
//		fmt.Println("Key b64 decode fail")
//	}
	fmt.Println("ds ", decKey)
//	fmt.Println("ds2 ", ds2)
//	key = kkk
	fmt.Println("kkk", len(kkk), kkk)      //  kkk is the real key saved from the encryption process and it works
	// data from store
/*	if _, err := os.Stat(string(id)); os.IsNotExist(err){
		// reply to client
		io.WriteString(w, `{"payload":"Data file not found"}`)
		return
	}
	encData, err := ioutil.ReadFile(string(id))
	if err != nil{
		// reply to client
		io.WriteString(w, `{"payload":"Error reading data file"}`)
	}
*/
	if store[string(id)] == nil{
		fmt.Println("ID not found in store")
		return
	}

	encData:= store[string(id)]


	var decData []byte
 	decData, err = decrypt(decKey, encData)

	if err != nil{
		log.Println("Decryption Error: ", err)
		return
	}

	// reply to client
	skey := base64.StdEncoding.EncodeToString(decData)
	s:=`{"payload":"`+skey+`"}`
	fmt.Println("s",s)
	io.WriteString(w, s)
}

func main() {
	http.HandleFunc("/store", StoreServer)
	http.HandleFunc("/retrieve", RetrieveServer)
	err := http.ListenAndServeTLS(":8443", "/etc/ssl/tmp/server.pem", "/etc/ssl/tmp/server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
