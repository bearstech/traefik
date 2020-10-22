package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
)

const fluentEndpoint = "localhost:0"

type Server struct {
	Blobs []interface{}
}

func New() *Server {
	return &Server{}
}

func (s *Server) ListenAndServe(address string, addr chan<- string, wg *sync.WaitGroup) error {
	listener, err := net.Listen("tcp", fluentEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	addr <- listener.Addr().String()

	fmt.Println(listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handler(conn, wg)
	}
}

func (s *Server) handler(conn net.Conn, wg *sync.WaitGroup) {
	defer conn.Close()
	decoder := msgpack.NewDecoder(conn)
	var m []interface{}

	for {
		blob, err := decoder.DecodeInterface()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("read error", err)
			return
		}

		var ok bool
		m, ok = blob.([]interface{})
		if !ok {
			log.Println("Not an array", blob)
			return
		}

		s.Blobs = append(s.Blobs, m)
		wg.Done()
	}
}

func TestFluent(t *testing.T) {

	SetFormatter(&logrus.TextFormatter{DisableColors: true})

	var wg sync.WaitGroup
	addr := make(chan string)
	srv := New()

	go srv.ListenAndServe(fluentEndpoint, addr, &wg)

	address := <-addr

	wg.Add(1)
	err := NewFluentHook(logrus.InfoLevel, fmt.Sprintf("http://%s", address), "test")
	if err != nil {
		t.Error(err)
	}

	wg.Add(1)
	ctx := context.Background()
	FromContext(ctx).Info("testing fluent")

	wg.Wait()

	// TODO: check how to validate
	for _, value := range srv.Blobs {
		fmt.Println(value)
	}

}
