package main

import (
	"io"
	"log"
	"net/http"
	"fmt"
	"io/ioutil"
)



// Commands to create certs
//
// openssl genrsa -out server.key 2048
// openssl ecparam -genkey -name secp384r1 -out server.key        
// openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
//
// use localhost for common name (CN) - this makes testing easier
//


func StoreServer(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("ReadAll: ", err)
	}
	// data to secure store
	fmt.Print(string(data))

	// reply to client
	io.WriteString(w, `{"a":1}`)
}

func RetrieveServer(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("ReadAll: ", err)
	}
	// data from secure store
	fmt.Print(string(data))

	// reply to client
	io.WriteString(w, `{"a":0}`)
}

func main() {
	http.HandleFunc("/store", StoreServer)
	http.HandleFunc("/retrieve", RetrieveServer)
	err := http.ListenAndServeTLS(":8443", "/tmp/server.pem", "/tmp/server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
