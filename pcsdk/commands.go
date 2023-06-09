package pcsdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/state"
)

type response struct {
	message string `json:"message"`
}

type Proxy struct {
	signing_key string
	url 		netip.AddrPort
}

func BuildProxy(control netip.AddrPort) Proxy {
	return Proxy {"", control}
}

func (p Proxy) Create(iport uint16, oport uint16, oip netip.Addr, id state.CustomUUID) error {
	_, err := p.execute(create(iport, oport, oip, id))
	return err
}

func (p Proxy) Modify(oport uint16, oip netip.Addr, id state.CustomUUID) error {
	_, err := p.execute(modify(oport, oip, id))
	return err
}

func (p Proxy) Delete(id state.CustomUUID) error {
	_, err := p.execute(delete(id))
	return err
}

// TODO: status function returning map of tunnels

func (p Proxy) execute(c command) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", errors.New(fmt.Sprintf("could not serialize: %s\n", err))
	}

	requestURL := fmt.Sprintf("http://%s:%d/command", p.url.Addr().String(), p.url.Port())
	fmt.Println(requestURL)
	bodyReader := bytes.NewReader(data)

	res, err := http.DefaultClient.Post(requestURL, "application/json", bodyReader)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error making http request: %s\n", err))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error reading response: %s\n", err))
	}

	if res.StatusCode != 202 && res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("error processing command: (%d) %s\n", res.StatusCode, body))
	} else {
		return string(body), nil
	}
}

type command struct {
	Create *commandCreate `json:"create,omitempty"`
	Modify *commandModify `json:"modify,omitempty"`
	Delete *commandDelete `json:"delete,omitempty"`
	Timestamp uint64	  `json:"timestamp,omitempty"`
	Signature string	  `json:"signature,omitempty"`
}

type commandCreate struct {
	IncomingPort    uint16     `json:"incoming_port"`
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	Id              string     `json:"id"`
}

func create(iport uint16, oport uint16, oip netip.Addr, id state.CustomUUID) command {
	cr:= commandCreate{iport, oport, oip, uuid.UUID.String(uuid.UUID(id))}
	c:= command{}
	c.Create = &cr
	return c
}

type commandModify struct {
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	Id              string     `json:"id"`
}

func modify(oport uint16, oip netip.Addr, id state.CustomUUID) command {
	m:= commandModify{oport, oip, uuid.UUID.String(uuid.UUID(id))}
	c:= command{}
	c.Modify = &m
	return c
}

type commandDelete struct {
	Id string `json:"id"`
}

func delete(id state.CustomUUID) command {
	d:= commandDelete{uuid.UUID.String(uuid.UUID(id))}
	c:= command{}
	c.Delete = &d
	return c
}