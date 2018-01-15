package client

import (
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type ClientStruct struct{}

func postJson(url string, content []byte) ([]byte, error) {

	// Need to know about self signed cert
	certs := x509.NewCertPool()
	// Read public key
	pemData, err := ioutil.ReadFile("/etc/myapp/ssl/tmp/server.pem")
	if err != nil {
		err := errors.New("Failed to find or read certificate")
		return nil, err
	}
	b := certs.AppendCertsFromPEM(pemData)
	if b == false {
		err := errors.New("Failed to load certificate")
		return nil, err
	}
	// Create a transport that knows server cert
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{RootCAs: certs},
		DisableCompression: true,
	}

	hClient := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", "https://localhost:8443/" + url, bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	// POST
	resp, err := hClient.Do(req)
	if err != nil {
		return nil, err
	}

	// get the body in a string
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	return data, err
}

func (ClientStruct) Store(id, payload []byte) ([]byte, error) {
	var (
		aesKey []byte
		err    error
	)

	data := map[string][]byte{"id": id, "data": payload}
	jData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Json Marshal fail")
		return nil, err
	}

	rData, err := postJson("store", jData)
	if err == nil {
		res := map[string][]byte{}
		json.Unmarshal(rData, &res)
		aesKey = res["aesKey"]
		if len(aesKey) == 0 {
			aesKey = res["error"]
		}
	}
	return aesKey, err
}

// Retrieve accepts an id and an AES key, and requests that the
// encryption-server retrieves the original (decrypted) bytes stored
// with the provided id
func (ClientStruct) Retrieve(id, aesKey []byte) ([]byte, error) {
	var (
		returnedData []byte
		err error
	)
	strAesKey := string(aesKey)
	if len(strAesKey) == 0 {
		err := errors.New("Key encode failed")
		return nil, err
	}
	data := map[string][]byte{"id": id, "aesKey": []byte(strAesKey)}
	jData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Json Marshal failed")
		return nil, err
	}

	rData, err := postJson("retrieve", jData)
	if err == nil {
		res := map[string][]byte{}
		err = json.Unmarshal(rData, &res)
		if err != nil {
			fmt.Println("Json UnMarshal fail")
			return nil, err
		}
		returnedErr := res["error"]
		returnedData = res["data"]
		if len(returnedErr)!= 0 {
			returnedData = res["error"]
		}
	}
	return returnedData, err
}
