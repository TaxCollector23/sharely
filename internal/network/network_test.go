package network

import (
	"fmt"
	"net"
	"testing"
)

func TestFreePortReturnsListenablePort(t *testing.T) {
	port, err := FreePort("127.0.0.1", 51234)
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("port %d reported free but could not be bound: %v", port, err)
	}
	l.Close()
}

func TestFreePortAvoidsCollision(t *testing.T) {
	p1, err := FreePort("127.0.0.1", 51300)
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p1))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	p2, err := FreePort("127.0.0.1", p1)
	if err != nil {
		t.Fatal(err)
	}
	if p2 == p1 {
		t.Fatalf("expected FreePort to skip the occupied port %d", p1)
	}
}

func TestReachable(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port
	if !Reachable("127.0.0.1", port) {
		t.Fatalf("expected listening port %d to be reachable", port)
	}
}
