package pcsdk

import (
	"encoding/json"
	"net/netip"
	"testing"

	"github.com/google/uuid"
	"github.com/thefeli73/polemos/state"
)

func TestCommandCreateJsonParse(t *testing.T) {
	ip, _ := netip.ParseAddr("127.0.0.99")
	uuid, _ := uuid.Parse("87e79cbc-6df6-4462-8412-85d6c473e3b1")
	c := NewCommandCreate(5555, 6666, ip, state.CustomUUID(uuid))
	m := ProxyCommandCreate{c}
	msg, err := json.Marshal(m)
	if err != nil {
		t.Fatalf(`%q`, err)
	}

	expected := "{\"create\":{\"incoming_port\":5555,\"destination_port\":6666,\"destination_ip\":\"127.0.0.99\",\"id\":\"87e79cbc-6df6-4462-8412-85d6c473e3b1\"}}"
	if string(msg) != expected {
		t.Fatalf(
			"\nExpected:\t %q\nGot:\t\t %q\n", expected, msg)
	}
}


