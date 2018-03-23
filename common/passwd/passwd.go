package passwd

import (
	"fmt"
	"github.com/piterparks/gopass"
	"os"
	"flag"
)

//get password from user input on console
func GetPwd()([]byte,error) {
	fmt.Println("please enter your password:")
	pwd,err := gopass.GetPasswd()
	if err != nil {
		return nil,err
	}
	return pwd,nil
}

//get password from user input and confire it.so user have to inout two times
func GetConfirePwd()([]byte,error){
	fmt.Println("please enter your password:")
	fpwd,err := gopass.GetPasswd()
	if err != nil{
		return nil,err
	}
	fmt.Println("please enter your password again:")
	spwd,err := gopass.GetPasswd()
	if err != nil{
		return nil,err
	}
	if len(fpwd) != len(spwd){
		fmt.Println("You have to enter the same password for twice!")
		os.Exit(1)
	}
	for i,v := range fpwd{
		if v != spwd[i]{
			fmt.Println("You have to enter the same password for twice!")
			os.Exit(1)
		}
	}
	return fpwd,nil
}

//get password where start node
func GetAccountPwd()([]byte,error){
	var pwd []byte
	var err error
	if len(os.Args) == 1{
		//start node with no params
		pwd,err = GetPwd()
		if err != nil {
			return nil, err
		}
	}else{
		var pstr string
		flag.StringVar(&pstr,"p","","wallet password")
		flag.Parse()
		if pstr == ""{
			fmt.Println("Invaild parameter, use '-p <password>' to specify a not nil wallet password.")
			os.Exit(1)
		}
		pwd = []byte(pstr)
	}
		return pwd,nil
}