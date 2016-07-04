//
// encrypt/decrypt application for core interview test
//

package main

import ("fmt";"flag";"../client"
	"strconv"
	"strings"
	"os"
	"io/ioutil"
	"bytes"
)

func usageText(){
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

func doOperation(op string, id []byte, param []byte, clientif client.Client)  {
	var err error
	var key []byte
	var data []byte
	if op == "e" {
		key, err = clientif.Store(id, param)
		if nil != err {
			fmt.Println(err.Error())
		} else {
			fmt.Print("Key:")
			fmt.Println(key)
			fmt.Println()
		}
	} else if op == "d" {
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

func main() {
	flag.Usage = usageText
	flag.Parse()
	args:=flag.Args()

	var in *os.File
	keyFileArg := args[2:]
	var keyFileStdin []byte
	if len(args) < 3 && in == nil{
		in = os.Stdin
		if in == nil {
			fmt.Println("Missing fields")
			usageText()
			return
		}
		keyFileStdin,_=ioutil.ReadAll(in)
	}

	p1:=[]byte(args[1])
	var p2 []byte

	if args[0]=="d" {
		offset := 0
		if len(args) >= 3 {
			// arg on cmd line
			keyFileArg[0] = strings.TrimLeft(keyFileArg[0], "Key: ")
			keyFileArg[0] = strings.TrimLeft(keyFileArg[0], "key: ")

			data:= make([]byte,len(args))
			for i, id := offset, 0; i < len(keyFileArg); i, id = i + 1, id + 1 {
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
			data:= make([]byte,len(keyFileStdin))
			for _,v := range spl {
				s:= v
				s= bytes.TrimLeft(s, "[")
				s= bytes.TrimRight(s, "]")
				if bytes.Contains(s, []byte("]")){
					for i,va:= range s {
						fmt.Println("rem",va, i)
						if va == byte(93){
							s = s[0:i]
						}
					}
				}
				ss := string(s)
				num, _ := strconv.Atoi(ss)
				data[i] = byte(num)
				i++
			}
			p2 = data[0:i]
		}
	}else{
		p2=[]byte(args[2])
	}

	var clientObj client.ClientStruct
	doOperation(args[0], p1, p2, clientObj)

}