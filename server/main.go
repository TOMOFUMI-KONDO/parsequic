package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pq "parsequic/proto"
)

var (
	port int
	cnt  = 0
)

func init() {
	flag.IntVar(&port, "port", 8080, "listen port")
	flag.Parse()
}

type server struct {
	pq.UnimplementedParseQuicServer
}

func newServer() *server {
	return &server{}
}

func (s *server) Parse(ctx context.Context, req *pq.ParseQuicRequest) (*pq.ParseQuicReply, error) {
	// TODO: parse quic using go-quic

	rep := &pq.ParseQuicReply{
		IsLongHeader: req.Data[0]&128 == 128,
	}

	return rep, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen; %v", err)
	}

	s := grpc.NewServer()
	pq.RegisterParseQuicServer(s, newServer())
	s.Serve(lis)
}
