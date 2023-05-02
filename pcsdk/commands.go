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

// ExecuteCommand  is the interface that wraps the Execute method.
type ExecuteCommand interface {
	Execute(netip.AddrPort) error
}

type response struct {
	message string `json:"message"`
}

// ProxyCommandStatus is the command to get the status of a proxy
type ProxyCommandStatus struct {
	Command CommandStatus `json:"status"`
	// signature Signature
}

// CommandStatus is the status of a proxy
type CommandStatus struct {}

// Execute is the method that executes the Status command
func (c ProxyCommandStatus) Execute(url netip.AddrPort) error {
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

	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("error processing command: (%d) %s\n", res.StatusCode, body))
	} else {
		return nil
	}
}

// NewCommandStatus returns a new CommandStatus
func NewCommandStatus() ProxyCommandStatus {
	c := CommandStatus{}
	return ProxyCommandStatus{c}
}

// ProxyCommandCreate is the command to create a proxy
type ProxyCommandCreate struct {
	Command CommandCreate `json:"create"`
	// signature Signature
}

// CommandCreate is the struct for the "Create" Command and contains all the info the proxy needs to create a tunnel
type CommandCreate struct {
	IncomingPort    uint16     `json:"incoming_port"`
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	Id              string     `json:"id"`
}

// Execute is the method that executes the Create command on a ProxyCommandCreate
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

// NewCommandCreate returns a new CommandCreate
func NewCommandCreate(iport uint16, oport uint16, oip netip.Addr, id state.CustomUUID) ProxyCommandCreate {
	c := CommandCreate{iport, oport, oip, uuid.UUID.String(uuid.UUID(id))}
	return ProxyCommandCreate{c}
}

// ProxyCommandModify is the command to modify a proxy
type ProxyCommandModify struct {
	Command CommandModify `json:"modify"`
}

// CommandModify is the struct for the "Modify" Command and contains all the info the proxy needs to modify a tunnel
type CommandModify struct {
	DestinationPort uint16     `json:"destination_port"`
	DestinationIP   netip.Addr `json:"destination_ip"`
	Id              string     `json:"id"`
}

// Execute is the method that executes the Modify command on a ProxyCommandModify
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

// NewCommandModify returns a new CommandModify
func NewCommandModify(oport uint16, oip netip.Addr, id state.CustomUUID) ProxyCommandModify {
	c := CommandModify{oport, oip, uuid.UUID.String(uuid.UUID(id))}
	return ProxyCommandModify{c}
}

// ProxyCommandDelete is the command to delete a proxy
type ProxyCommandDelete struct {
	Command CommandDelete `json:"delete"`
}

// CommandDelete is the struct for the "Delete" Command and contains the id of the tunnel to delete
type CommandDelete struct {
	Id string `json:"id"`
}

// Execute is the method that executes the Delete command on a ProxyCommandDelete
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

// NewCommandDelete returns a new CommandDelete
func NewCommandDelete(id state.CustomUUID) ProxyCommandDelete {
	c := CommandDelete{uuid.UUID.String(uuid.UUID(id))}
	return ProxyCommandDelete{c}
}
