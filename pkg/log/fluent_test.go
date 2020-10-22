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
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

const fluentEndpoint = "localhost:0"

type Message []interface{}

type Server struct {
	Messages []Message
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

		s.Messages = append(s.Messages, m)
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
	err := AddFluentHook(logrus.InfoLevel, fmt.Sprintf("http://%s", address), "test")
	if err != nil {
		t.Error(err)
	}

	wg.Add(1)
	ctx := context.Background()
	FromContext(ctx).Info("testing fluent")

	wg.Wait()

	values, ok := srv.Messages[1][2].(map[string]interface{})
	if !ok {
		t.Error("Can't unpack message pack fluent values")
	}

	assert.Equal(t, values[""].(string), "testing fluent")

}
