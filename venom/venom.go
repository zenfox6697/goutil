package venom

import (
	"encoding/hex"
	"fmt"

	shellcode "github.com/brimstone/go-shellcode"
)

type venom string

// hex shellcode or const preset
func NewVenom(sc string) venom {
	return venom(sc)
}

func (v venom) Run() {
	b, err := hex.DecodeString(string(v))
	if err != nil {
		fmt.Println("")
		return
	}
	shellcode.Run(b)
}
