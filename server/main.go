package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/lucas-clemente/quic-go/external/protocol"

	pq "parsequic/proto"

	"github.com/lucas-clemente/quic-go/external/wire"
	"google.golang.org/grpc"
)

var (
	port int
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
	hdr, _, _, err := wire.ParsePacket(req.Data, 4)
	if err != nil {
		fmt.Printf("failed to ParsePacket(); %v\n", err)
		return &pq.ParseQuicReply{}, err
	}

	rep := &pq.ParseQuicReply{
		IsLongHeader: hdr.IsLongHeader,
		Type:         packetType(hdr.Type),
		Version:      uint32(hdr.Version),
		DstConnID:    hdr.DestConnectionID,
		SrcConnID:    hdr.SrcConnectionID,
	}
	fmt.Printf("isLongHeader:%t type:%s version:%d dstConnID:%x srcConnID:%x\n",
		rep.IsLongHeader,
		rep.Type,
		rep.Version,
		rep.DstConnID,
		rep.SrcConnID,
	)

	return rep, nil
}

func packetType(pt protocol.PacketType) pq.PacketType {
	if pt == protocol.PacketTypeInitial {
		return pq.PacketType_INITIAL
	} else if pt == protocol.PacketType0RTT {
		return pq.PacketType_ZERO_RTT
	} else if pt == protocol.PacketTypeHandshake {
		return pq.PacketType_HANDSHAKE
	} else if pt == protocol.PacketTypeRetry {
		return pq.PacketType_RETRY
	} else {
		return pq.PacketType_SHORT_HEADER
	}
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
