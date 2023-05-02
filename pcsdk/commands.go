package pcsdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/netip"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/state"
)

type ExecuteCommand interface {
	Execute(netip.AddrPort) error
}

type response struct {
	message string `json:"message"`
}

type ProxyCommandCreate struct {
	Command CommandCreate `json:"create"`
	// signature Signature
}

type CommandCreate struct {
	IncomingPort    uint16     `json:"incoming_port"`
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	Id              string     `json:"id"`
}

func (c ProxyCommandCreate) Execute(url netip.AddrPort) error {
	data, err := json.Marshal(c)
	if err != nil {
		return errors.New(fmt.Sprintf("could not serialize: %s\n", err))
	}

	requestURL := fmt.Sprintf("http://%s:%d/command", url.Addr().String(), url.Port())
	bodyReader := bytes.NewReader(data)

	res, err := http.DefaultClient.Post(requestURL, "application/json", bodyReader)
	if err != nil {
		return errors.New(fmt.Sprintf("error making http request: %s\n", err))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading response: %s\n", err))
	}

	if res.StatusCode != 202 {
		return errors.New(fmt.Sprintf("error processing command: (%d) %s\n", res.StatusCode, body))
	} else {
		return nil
	}
}

func NewCommandCreate(iport uint16, oport uint16, oip netip.Addr, id state.CustomUUID) ProxyCommandCreate {
	c := CommandCreate{iport, oport, oip, uuid.UUID.String(uuid.UUID(id))}
	return ProxyCommandCreate{c}
}

type ProxyCommandModify struct {
	Command CommandModify `json:"modify"`
}

type CommandModify struct {
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	Id              string     `json:"id"`
}

func (c ProxyCommandModify) Execute(url netip.AddrPort) error {
	data, err := json.Marshal(c)
	if err != nil {
		return errors.New(fmt.Sprintf("could not serialize: %s\n", err))
	}

	requestURL := fmt.Sprintf("http://%s:%d/command", url.Addr().String(), url.Port())

	bodyReader := bytes.NewReader(data)

	res, err := http.DefaultClient.Post(requestURL, "application/json", bodyReader)
	if err != nil {
		return errors.New(fmt.Sprintf("error making http request: %s\n", err))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading response: %s\n", err))
	}

	if res.StatusCode != 202 {
		return errors.New(fmt.Sprintf("error processing command: (%d) %s\n", res.StatusCode, body))
	} else {
		return nil
	}
}

func NewCommandModify(oport uint16, oip netip.Addr, id state.CustomUUID) ProxyCommandModify {
	c := CommandModify{oport, oip, uuid.UUID.String(uuid.UUID(id))}
	return ProxyCommandModify{c}
}

type ProxyCommandDelete struct {
	Command CommandDelete `json:"delete"`
}

type CommandDelete struct {
	Id string `json:"id"`
}

func (c ProxyCommandDelete) Execute(url netip.AddrPort) error {
	data, err := json.Marshal(c)
	if err != nil {
		return errors.New(fmt.Sprintf("could not serialize: %s\n", err))
	}

	requestURL := fmt.Sprintf("http://%s:%d/command", url.Addr().String(), url.Port())

	bodyReader := bytes.NewReader(data)

	res, err := http.DefaultClient.Post(requestURL, "application/json", bodyReader)
	if err != nil {
		return errors.New(fmt.Sprintf("error making http request: %s\n", err))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading response: %s\n", err))
	}

	if res.StatusCode != 202 {
		return errors.New(fmt.Sprintf("error processing command: (%d) %s\n", res.StatusCode, body))
	} else {
		return nil
	}
}

func NewCommandDelete(id state.CustomUUID) ProxyCommandDelete {
	c := CommandDelete{uuid.UUID.String(uuid.UUID(id))}
	return ProxyCommandDelete{c}
}
