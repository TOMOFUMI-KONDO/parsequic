package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	pq "parsequic/proto"

	"google.golang.org/grpc"

	"github.com/lucas-clemente/quic-go/protocol"
	"github.com/lucas-clemente/quic-go/wire"
)

var (
	port string
)

func init() {
	flag.StringVar(&port, "port", ":8080", "listen port")
	flag.Parse()
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(port))
	if err != nil {
		log.Fatalf("failed to listen; %v\n", err)
	}

	s := grpc.NewServer()
	pq.RegisterParseQuicServer(s, newServer())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}

type server struct {
	pq.UnimplementedParseQuicServer
}

func newServer() *server {
	return &server{}
}

func (s *server) Parse(req *pq.ParseQuicRequest, stream pq.ParseQuic_ParseServer) error {
	data := req.GetData()

	for {
		hdr, _, rest, err := wire.ParsePacket(data, 4) // 4 is sloppy
		if err != nil {
			return err
		}

		go func() {
			var pt pq.PacketType
			if hdr.IsLongHeader {
				pt = longHeaderPacketType(hdr.Type)
			} else {
				pt = pq.PacketType_ONE_RTT
			}

			stream.Send(&pq.ParseQuicReply{
				IsLongHeader: hdr.IsLongHeader,
				Type:         pt,
				Version:      uint32(hdr.Version),
				DstConnID:    hdr.DestConnectionID,
				SrcConnID:    hdr.SrcConnectionID,
			})
		}()

		if rest == nil {
			break
		}
		data = rest
	}

	return nil
}

func longHeaderPacketType(pt protocol.PacketType) pq.PacketType {
	if pt == protocol.PacketTypeInitial {
		return pq.PacketType_INITIAL
	} else if pt == protocol.PacketType0RTT {
		return pq.PacketType_ZERO_RTT
	} else if pt == protocol.PacketTypeHandshake {
		return pq.PacketType_HANDSHAKE
	} else if pt == protocol.PacketTypeRetry {
		return pq.PacketType_RETRY
	} else {
		return pq.PacketType_VERSION_NEGOTIATION
	}
}
