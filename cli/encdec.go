//
// encrypt/decrypt application for core interview test
//

package main

import ("fmt";"flag";"../client"
//	"encoding/base64"
//	"encoding/base64"
	"strconv"
)

func usageText(){
	fmt.Println("Encrypt-Decrypt")
	fmt.Println("   encdec e id \"data\"")
	fmt.Println("       e to encrypt")
	fmt.Println("       id to identify data")
	fmt.Println("       data to encrypt")
	fmt.Println("    this will return the key required to decrypt the data")

	fmt.Println("")
	fmt.Println("   encdec d id key")
	fmt.Println("       d to decrypt")
	fmt.Println("       id to identify data")
	fmt.Println("       key to unlock data")
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
			fmt.Println("Key:")
			for i:=0; i<len(key); i++{
				fmt.Print(key[i], " ")
			}
			fmt.Println()
			fmt.Println("Key:")
			fmt.Println(key)
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

	if len(args) < 3{
		fmt.Println("Missing fields")
		usageText()
		return
	}

	p1:=[]byte(args[1])
	var p2 []byte

	data:= make([]byte,len(args))

	if args[0]=="d" {
		for i := 2; i < len(args); i++ {
			s := (args[i])
			v,_:=strconv.Atoi(s)
			data[i-2] = byte(v)
		}
		p2=data
	}else{
		p2=[]byte(args[2])
	}
//	fmt.Println(args)
	fmt.Println(p2)

	var clientObj client.ClientStruct

	doOperation(args[0], p1, p2, clientObj)

}