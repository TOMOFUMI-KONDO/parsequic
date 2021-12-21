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

	var pktType pq.LongHeaderPacketType
	if hdr.Type == protocol.PacketTypeInitial {
		pktType = pq.LongHeaderPacketType_INITIAL
	} else if hdr.Type == protocol.PacketType0RTT {
		pktType = pq.LongHeaderPacketType_ZERO_RTT
	} else if hdr.Type == protocol.PacketTypeHandshake {
		pktType = pq.LongHeaderPacketType_HANDSHAKE
	} else if hdr.Type == protocol.PacketTypeRetry {
		pktType = pq.LongHeaderPacketType_RETRY
	}

	rep := &pq.ParseQuicReply{
		IsLongHeader: hdr.IsLongHeader,
		Type:         pktType,
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

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen; %v", err)
	}

	s := grpc.NewServer()
	pq.RegisterParseQuicServer(s, newServer())
	s.Serve(lis)
}
