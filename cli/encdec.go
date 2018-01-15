//
// encrypt/decrypt application for core interview test
//

package main

import (
	"fmt"
	"flag"
	"../client"
	"strconv"
	"strings"
	"os"
	"io/ioutil"
	"bytes"
//	"image/draw"
)

func usageText() {
	fmt.Println("Encrypt-Decrypt")
	fmt.Println("   encdec e id \"data\"")
	fmt.Println("       e to encrypt")
	fmt.Println("       id to identify data")
	fmt.Println("       data to encrypt")
	fmt.Println("    this will return the key required to decrypt the data")
	fmt.Println("	encdec e id \"data\" > key.txt")

	fmt.Println("")
	fmt.Println("   encdec d id key")
	fmt.Println("       d to decrypt")
	fmt.Println("       id to identify data")
	fmt.Println("       key to unlock data")
	fmt.Println("	encdec d id < key.txt")
	fmt.Println("	encdec d id Key :[xx xx xx xx xx...]")
	fmt.Println("")
}

func doOperation(op string, id []byte, param []byte, clientif client.Client) {
	var (
		err  error
		key  []byte
		data []byte
	)

	switch {
		case op == "e" : {
			key, err = clientif.Store(id, param)
			if nil != err {
				fmt.Println(err.Error())
			} else {
				fmt.Print("Key:")
				fmt.Println(key)
				fmt.Println()
			}
		}
		case op == "d" : {
//			fmt.Println(param)
			data, err = clientif.Retrieve(id, param)
			if nil != err {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Received from server:")
				fmt.Println(string(data))
			}
		}
	}
}

func main() {
	flag.Usage = usageText
	flag.Parse()
	args := flag.Args()

	var in *os.File
	in = os.Stdin
	if len(args) < 2 || in == nil {
		fmt.Println("Missing fields")
		usageText()
		return
	}
	operation := args[0]

	var (
		keyFileStdin []byte
		err error
	)
	switch {
	case len(args) == 2 && operation == "e" :
		fmt.Println("Field missing")
		usageText()
		return
	case len(args) == 2 && operation == "d" :
		// key data from redirected file, this doesnt work when the file is specified in the debugger
		keyFileStdin, err = ioutil.ReadAll(in)
		if err != nil {
			fmt.Println("Error in reading key file data")
			return
		}
	case len(args) > 5 && operation == "e" :
		fmt.Println("Too many fields")
		usageText()
		return
	}

	id := []byte(args[1])
	var p2 []byte

	switch {
	case operation == "e": // param for encode operation
		p2 = []byte(args[2])
	case operation == "d": // param for decode operation
		if len(args) >= 3 && args[2] != "<" {
			// arg on cmd line
			keyFileArg := args[2:]
			if len(keyFileArg) == 0 {
				fmt.Println("Parameter field error")
				return
			}
			keyFileArg[0] = strings.TrimLeft(keyFileArg[0], "Key: ")
			keyFileArg[0] = strings.TrimLeft(keyFileArg[0], "key: ")

			data := make([]byte, len(args))
			for i:= 0; i < len(keyFileArg); i = i+1 {
				s := keyFileArg[i]
				s = strings.TrimLeft(s, "[ ")
				s = strings.TrimRight(s, "] ")
				v, err := strconv.Atoi(s)
				if err != nil {
					fmt.Println("Error in keyfile")
					return
				}
				data[i] = byte(v)
			}
			p2 = data
		} else {
			// arg from redirected from stdin
			keyFileStdin = bytes.TrimLeft(keyFileStdin, "Key: ")
			keyFileStdin = bytes.TrimLeft(keyFileStdin, "key: ")
			spl := bytes.Split(keyFileStdin, []byte(" "))
			i := 0
			data := make([]byte, len(keyFileStdin))
			for _, v := range spl {
				s := v
				s = bytes.TrimLeft(s, "[")
				s = bytes.TrimRight(s, "]")
				if bytes.Contains(s, []byte("]")) {
					for i, va := range s {
						if va == byte(93) {
							s = s[0:i]
						}
					}
				}
				ss := string(s)
				num, err := strconv.Atoi(ss)
				if err != nil {
					fmt.Println("Atoi fail ", ss)
					return
				}
				data[i] = byte(num)
				i++
			}
			p2 = data[0:i]
		}
	default:
		fmt.Println("Unknown command.")
		usageText()
		return
	}


	// create interface to connect client
	var clientObj client.ClientStruct
	doOperation(operation, id, p2, clientObj)

}
