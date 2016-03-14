package transport_test

import (
	"strings"
	"testing"

	"github.com/micro/go-micro/transport"
)

func expectedPort(t *testing.T, expected string, lsn transport.Listener) {
	parts := strings.Split(lsn.Addr(), ":")
	port := parts[len(parts)-1]

	if port != expected {
		lsn.Close()
		t.Errorf("Expected address to be `%s`, got `%s`", expected, port)
	}
}

func TestHTTPTransportPortRange(t *testing.T) {
	tp := transport.NewTransport([]string{})

	lsn1, err := tp.Listen(":44444-44448")
	if err != nil {
		t.Errorf("Did not expect an error, got %s", err)
	}
	expectedPort(t, "44444", lsn1)

	lsn2, err := tp.Listen(":44444-44448")
	if err != nil {
		t.Errorf("Did not expect an error, got %s", err)
	}
	expectedPort(t, "44445", lsn2)

	lsn, err := tp.Listen(":0")
	if err != nil {
		t.Errorf("Did not expect an error, got %s", err)
	}

	lsn.Close()
	lsn1.Close()
	lsn2.Close()
}

func TestHTTPTransportCommunication(t *testing.T) {
	tr := transport.NewTransport([]string{})

	l, err := tr.Listen(":0")
	if err != nil {
		t.Errorf("Unexpected listen err: %v", err)
	}
	defer l.Close()

	fn := func(sock transport.Socket) {
		for {
			var m transport.Message
			if err := sock.Recv(&m); err != nil {
				return
			}

			t.Logf("Successfully received %+v", m)

			if err := sock.Send(&m); err != nil {
				return
			}
		}
	}

	go func() {
		if err := l.Accept(fn); err != nil {
			t.Errorf("Unexpected accept err: %v", err)
		}
	}()

	c, err := tr.Dial(l.Addr())
	if err != nil {
		t.Errorf("Unexpected dial err: %v", err)
	}
	defer c.Close()

	m := transport.Message{
		Header: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "Hello World"}`),
	}

	if err := c.Send(&m); err != nil {
		t.Errorf("Unexpected send err: %v", err)
	}

	var rm transport.Message

	if err := c.Recv(&rm); err != nil {
		t.Errorf("Unexpected recv err: %v", err)
	}

	if string(rm.Body) != string(m.Body) {
		t.Errorf("Expected %v, got %v", m.Body, rm.Body)
	}
}
