package svm

type Opcode byte

const(
	PSH Opcode = iota
	ADD
	POP
	SET
	HLT
)
