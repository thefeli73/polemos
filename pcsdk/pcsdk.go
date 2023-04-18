package pcsdk

import (
	"net/netip"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/state"
)

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

func NewCommandCreate(iport uint16, oport uint16, oip netip.Addr, id state.CustomUUID) CommandCreate {
	return CommandCreate{iport, oport, oip, uuid.UUID.String(uuid.UUID(id))}
}
