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
			fmt.Println(param)
			data, err = clientif.Retrieve(id, param)
			if nil != err {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Data:")
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
	if len(args) < 3 || in == nil {
		fmt.Println("Missing fields")
		usageText()
		return
	}
	operation := args[1]
	keyFileArg := args[2:]
	var keyFileStdin []byte
	var err error
	//	if len(args) < 3 && in == nil
	//	{
	//		if in == nil {
	//			fmt.Println("Missing fields")
	//			usageText()
	//			return
	//		}
	if len(args) == 3 {
		keyFileStdin, err = ioutil.ReadAll(in)
		if err != nil {
			fmt.Println("Error in reading key file ")
			return
		}
	}

	id := []byte(args[2])
	var p2 []byte

	switch {
	case operation == "d": // param for decode operation
		offset := 0
		if len(args) >= 3 {
			// arg on cmd line
			keyFileArg[0] = strings.TrimLeft(keyFileArg[0], "Key: ")
			keyFileArg[0] = strings.TrimLeft(keyFileArg[0], "key: ")

			data := make([]byte, len(args))
			for i, id := offset, 0; i < len(keyFileArg); i, id = i+1, id+1 {
				s := (keyFileArg[i])
				s = strings.TrimLeft(s, "[ ")
				s = strings.TrimRight(s, "] ")
				v, _ := strconv.Atoi(s)
				data[id] = byte(v)
			}
			p2 = data
		} else {
			// arg from stdin
			if bytes.Equal(keyFileStdin[0:6], ([]byte("Key:\n[")[:])) == true {
				offset = 6
			}
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
						fmt.Println("rem", va, i)
						if va == byte(93) {
							s = s[0:i]
						}
					}
				}
				ss := string(s)
				num, err := strconv.Atoi(ss)
				if err != nil {
					fmt.Println("Atoi fail ", ss)

				}
				data[i] = byte(num)
				i++
			}
			p2 = data[0:i]
		}
	case operation == "e": // param for encode operation
		p2 = []byte(args[3])
	default:
		fmt.Println("Unknown command.")
		usageText()
		return
	}

	// create interface to connect client
	var clientObj client.ClientStruct
	doOperation(operation, id, p2, clientObj)

}
