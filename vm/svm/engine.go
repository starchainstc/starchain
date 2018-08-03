package svm

import "fmt"

type Engine struct{
	Code []byte
	Ip uint64
	Sp uint64
	Stack []byte
}

func (e *Engine)Execute() error{

}

func (e *Engine)eval(instr Opcode){
	switch instr{
	case ADD:
		a := e.Stack[e.Sp]
		e.Sp--
		b := e.Stack[e.Sp]
		e.Sp--
		a = a+b
		e.Stack[e.Sp] = a
		e.Sp++
		fmt.Println("add")
	case POP:
		e.Stack[e.Sp] = 0
		e.Sp--
		fmt.Println("pop")
	case PSH:
		//读取操作数
		e.Ip++
		e.Stack[e.Sp] = e.Code[e.Ip]
		//栈指针向右移动
		e.Sp++
		fmt.Println("psh")
	case SET:
		fmt.Println("set")
	case HLT:
		fmt.Println("hlt")
	default:
		fmt.Println("no code")
	}
}
