package client

import (
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
	"bytes"
	"encoding/json"
	"errors"
	"encoding/base64"
)

type ClientStruct struct{}

func postJson(url string, content []byte) ([]byte, error) {

	// Need to know about self signed cert
	certs := x509.NewCertPool()
	// Read public key
	pemData, err := ioutil.ReadFile("/etc/ssl/tmp/server.pem")
	if err != nil {
		return nil, err
	}
	b := certs.AppendCertsFromPEM(pemData)
	if b == false {
		e := errors.New("Failed to load certificate")
		return nil, e
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
	var aesKey []byte
	var err error

	data := map[string][]byte{"id": id, "payload": payload}
	jData, _ := json.Marshal(data)
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
	var payload []byte
	var err error
	strAesKey := base64.StdEncoding.EncodeToString(aesKey)
	if len(strAesKey) == 0 {
		err := errors.New("Key encode failed")
		return nil, err
	}
	data := map[string][]byte{"id": id, "aesKey": []byte(strAesKey)}
	jData, _ := json.Marshal(data)

	rData, err := postJson("retrieve", jData)

	if err == nil {
		res := map[string][]byte{}
		json.Unmarshal(rData, &res)
		payload = res["payload"]
		if len(payload) == 0 {
			payload = res["error"]
		}
	}
	return payload, err
}
