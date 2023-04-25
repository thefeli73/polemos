package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/pcsdk"
	"github.com/thefeli73/polemos/state"
)

func main() {
	uuid := uuid.MustParse("87e79cbc-6df6-4462-8412-85d6c473e3b1")

	m := pcsdk.NewCommandDelete(state.CustomUUID(uuid))
	err := m.Execute("http://localhost:3000")
	if err != nil {
		fmt.Printf("error executing modify command: %s\n", err)
	} else {
		fmt.Println("executing modify command completed")
	}
}
