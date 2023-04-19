package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/netip"
	"os"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/pcsdk"
	"github.com/thefeli73/polemos/state"
)

func main() {
	ip := netip.MustParseAddr("127.0.0.1")
	uuid := uuid.MustParse("87e79cbc-6df6-4462-8412-85d6c473e3b1")

	m := pcsdk.NewCommandCreate(5555, 6666, ip, state.CustomUUID(uuid))
	data, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("client: could not serialize into JSON")
		os.Exit(1)
	}

	fmt.Printf(string(data))

	requestURL := "http://localhost:3000/command"
	bodyReader := bytes.NewReader(data)
	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)

	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	fmt.Println(req)
	res, err := http.DefaultClient.Post(requestURL, "application/json", bodyReader)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: %s\n", resBody)
}
