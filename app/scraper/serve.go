package scraper

import (
	"encoding/hex"
	"log"
	"net"

	pb "github.com/DistributedSolutions/LendingBot/app/scraper/scraperGRPC"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type IScraper interface {
	GetLastDayAndSecond() (day []byte, second []byte, err error)
	LoadDay(day []byte) error
	ReadNext() ([]byte, error)

	// Not Implmeneted over GRPC
	SetDay(day []byte)
	SetSecond(second []byte)
	LoadSecond(second []byte) ([]byte, error)
	ReadLast() ([]byte, error)
}

type GRPSScraper struct {
	Sc *Scraper
}

// SayHello implements helloworld.GreeterServer
/*func (s *GRPSScraper) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}*/

func (sc *GRPSScraper) GetLastDayAndSecond(ctx context.Context, in *pb.Empty) (*pb.LoadLastReply, error) {
	day, second, err := sc.Sc.Walker.GetLastDayAndSecond()
	return &pb.LoadLastReply{Day: hex.EncodeToString(day), Second: hex.EncodeToString(second)}, err
}

func (sc *GRPSScraper) LoadDay(ctx context.Context, in *pb.Message) (*pb.Empty, error) {
	hexStr := in.Message
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}

	err = sc.Sc.Walker.LoadDay(data)
	return nil, err
}

func (sc *GRPSScraper) ReadNext(ctx context.Context, in *pb.Empty) (*pb.Message, error) {
	next, err := sc.Sc.Walker.ReadNext()
	if err != nil {
		return nil, err
	}

	m := new(pb.Message)
	pb.Message = hex.EncodeToString(next)
	return m, nil
}

func (sc *Scraper) Serve() {
	gsc := new(GRPSScraper)
	gsc.Sc = sc

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterScraperGRPCServer(s, gsc)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

/*func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}*/
