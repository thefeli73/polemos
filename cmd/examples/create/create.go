package main

import (
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/pcsdk"
	"github.com/thefeli73/polemos/state"
)

func main() {
	proxy := pcsdk.BuildProxy(netip.MustParseAddrPort("127.0.0.1:14000"))
	ip := netip.MustParseAddr("127.0.0.1")
	uuid := uuid.MustParse("87e79cbc-6df6-4462-8412-85d6c473e3b1")

	err := proxy.Create(5555, 8080, ip, state.CustomUUID(uuid))
	if err != nil {
		fmt.Printf("error executing create command: %s\n", err)
	} else {
		fmt.Println("executing create command completed")
	}
}
