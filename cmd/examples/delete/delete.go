package main

import (
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/pcsdk"
	"github.com/thefeli73/polemos/state"
)

func main() {
	uuid := uuid.MustParse("87e79cbc-6df6-4462-8412-85d6c473e3b1")

	m := pcsdk.NewCommandDelete(state.CustomUUID(uuid))
	err := m.Execute(netip.MustParseAddrPort("127.0.0.1:3000"))
	if err != nil {
		fmt.Printf("error executing delete command: %s\n", err)
	} else {
		fmt.Println("executing delete command completed")
	}
}
