package client

import (
	"log"
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
	"bytes"
	"encoding/json"
)

// not used
func get(url string) string {

	// Need to know about self signed cert
	certs := x509.NewCertPool()
	pemData, err := ioutil.ReadFile("/tmp/server.pem")
	if err != nil {
		// do error
	}
	certs.AppendCertsFromPEM(pemData)

	// Create a transport that knows server cert
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{RootCAs: certs},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	// Get
	resp, err := client.Get("https://localhost:8443/"+url)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	// All this to get the body in a string
	defer resp.Body.Close()
	data,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ReadAll: ", err)
	}
	return string(data)
}


func postjson(url string, content []byte ) ([]byte, err error) {

	// Need to know about self signed cert
	certs := x509.NewCertPool()
	// Read public key
	pemData, err := ioutil.ReadFile("/tmp/server.pem")
	if err != nil {
		return nil, err
	}
	certs.AppendCertsFromPEM(pemData)

	// Create a transport that knows server cert
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{RootCAs: certs},
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", "https://localhost:8443/"+url, bytes.NewBuffer(content))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	// POST
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// All this to get the body in a string
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	return data, err
}

func Store(id, payload []byte) ([]byte, error) {
	var aesKey []byte
	var err error
	data :=  map[string]int{"id": id, "payload": payload}
	jdata, _ := json.Marshal(data)

	rdata, err := postjson("store", jdata)

	if err == nil {
		res := map[string]int{}
		json.Unmarshal(rdata, &res)
		aesKey = res["aesKey"]
	}
	return aesKey, err
}

// Retrieve accepts an id and an AES key, and requests that the
// encryption-server retrieves the original (decrypted) bytes stored
// with the provided id
func Retrieve(id, aesKey []byte) ([]byte, error) {
	var payload []byte
	var err error
	data :=  map[string]int{"id": id, "aesKey": aesKey}
	jdata, _ := json.Marshal(data)

	rdata, err := postjson("retrieve", jdata)

	if err == nil {
		res := map[string]int{}
		json.Unmarshal(rdata, &res)
		payload = res["payload"]
	}

	return payload, err
}
