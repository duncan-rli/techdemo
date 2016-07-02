package main

import ("fmt";"flag";"encdec/client"
//	"encoding/base64"
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
		if 0 == len(err.Error()) {
			fmt.Println(err.Error())
		} else {
			fmt.Println(key)
		}
	} else if op == "d" {
		data, err = clientif.Retrieve(id, param)
		if 0 == len(err.Error()) {
			fmt.Println(err.Error())
		} else {
			fmt.Println(data)
		}
	}
}

func main() {
	flag.Usage = usageText
	flag.Parse()
	args:=flag.Args()

	if len(args) == 0{
		fmt.Println("Missing fields")
		usageText()
		return
	}
	if len(args) > 3{
		fmt.Println("Too many fields")
		usageText()
		return
	}

fmt.Println(args[0], args[1], args[2])

	// will probably need to encode things better like this
//	p1:= base64.StdEncoding(args[1])
//	p2:= base64.StdEncoding(args[2])

	p1:=[]byte(args[1])
	p2:=[]byte(args[2])
	var encdecClient client.Client
	doOperation(args[0], p1, p2, encdecClient)


}