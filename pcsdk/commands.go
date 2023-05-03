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

// Proxy owns the interfaces for the pcsdk.
type Proxy struct {
	signing_key string
	url 		netip.AddrPort
}

// BuildProxy creates a proxy struct for the given url to easily interact with that proxy instance (create, edit, delete tunnels etc)
func BuildProxy(proxyIP netip.AddrPort) Proxy {
	return Proxy {"", proxyIP}
}

// Create a tunnel with the given parameters.
func (p Proxy) Create(entryPort uint16, servicePort uint16, serviceIP netip.Addr, serviceUUID state.CustomUUID) error {
	_, err := p.execute(create(entryPort, servicePort, serviceIP, serviceUUID))
	return err
}

// Modify a tunnel with the given parameters.
func (p Proxy) Modify(servicePort uint16, serviceIP netip.Addr, serviceUUID state.CustomUUID) error {
	_, err := p.execute(modify(servicePort, serviceIP, serviceUUID))
	return err
}

// Delete a tunnel with the given parameters.
func (p Proxy) Delete(serviceUUID state.CustomUUID) error {
	_, err := p.execute(delete(serviceUUID))
	return err
}

// TODO: status function returning map of tunnels
// Status returns a list of tunnels for the given proxy.
func (p Proxy) Status() error {
	_, err := p.execute(status())
	return err
}


func (p Proxy) execute(c command) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", errors.New(fmt.Sprintf("could not serialize: %s\n", err))
	}

	requestURL := fmt.Sprintf("http://%s:%d/command", p.url.Addr().String(), p.url.Port())
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
	Status *commandStatus `json:"status,omitempty"`
	Timestamp uint64	  `json:"timestamp,omitempty"`
	Signature string	  `json:"signature,omitempty"`
}

type commandCreate struct {
	IncomingPort    uint16     `json:"incoming_port"`
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	ID              string     `json:"id"`
}

func create(entryPort uint16, servicePort uint16, serviceIP netip.Addr, serviceUUID state.CustomUUID) command {
	cr:= commandCreate{entryPort, servicePort, serviceIP, uuid.UUID.String(uuid.UUID(serviceUUID))}
	c:= command{}
	c.Create = &cr
	return c
}

type commandModify struct {
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	ID              string     `json:"id"`
}

func modify(servicePort uint16, serviceIP netip.Addr, serviceUUID state.CustomUUID) command {
	m:= commandModify{servicePort, serviceIP, uuid.UUID.String(uuid.UUID(serviceUUID))}
	c:= command{}
	c.Modify = &m
	return c
}

type commandDelete struct {
	ID string `json:"id"`
}

func delete(serviceUUID state.CustomUUID) command {
	d:= commandDelete{uuid.UUID.String(uuid.UUID(serviceUUID))}
	c:= command{}
	c.Delete = &d
	return c
}


type commandStatus struct {
}

func status() command {
	d:= commandStatus{}
	c:= command{}
	c.Status = &d
	return c
}
