package client

import (
	"encoding/hex"
	"log"
	// "os"

	pb "github.com/Emyrk/LendingBot/src/scraper/scraperGRPC"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

/*
type IScraper interface {
	SetDay(day []byte)
	SetSecond(second []byte)
	GetLastDayAndSecond() (day []byte, second []byte, err error)
	LoadDay(day []byte) error
	LoadSecond(second []byte) ([]byte, error)
	ReadLast() ([]byte, error)
}

*/

const (
	address     = "localhost:50051"
	defaultName = "world"
)

type ScraperClient struct {
	Address string
	Name    string
	Client  pb.ScraperGRPCClient
	Conn    *grpc.ClientConn
}

func NewScraperClient(name string, add string) *ScraperClient {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewScraperGRPCClient(conn)

	sc := new(ScraperClient)
	sc.Client = c
	sc.Name = name
	sc.Address = add
	sc.Conn = conn

	return sc
}

func (sc *ScraperClient) Close() {
	sc.Conn.Close()
}

func (sc *ScraperClient) GetLastDayAndSecond() (day []byte, second []byte, err error) {
	ret, err := sc.Client.GetLastDayAndSecond(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, nil, err
	}

	day, err = hex.DecodeString(ret.Day)
	if err != nil {
		return nil, nil, err
	}

	second, err = hex.DecodeString(ret.Second)
	if err != nil {
		return nil, nil, err
	}

	return
}

func (sc *ScraperClient) LoadDay(data []byte) (err error) {
	m := &pb.Message{Message: hex.EncodeToString(data)}
	_, err = sc.Client.LoadDay(context.Background(), m)
	return
}

func (sc *ScraperClient) ReadNext() (data []byte, err error) {
	ret, err := sc.Client.ReadNext(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}

	data, err = hex.DecodeString(ret.Message)
	if err != nil {
		return nil, err
	}

	return
}

// func main() {
// 	sc := NewScraperClient(defaultName, address)
// 	defer sc.Close()

// 	d, s, e := sc.GetLastDayAndSecond()
// 	log.Println(d, s, e)

// 		// Set up a connection to the server.
// 		conn, err := grpc.Dial(address, grpc.WithInsecure())
// 		if err != nil {
// 			log.Fatalf("did not connect: %v", err)
// 		}
// 		defer conn.Close()
// 		c := pb.NewGreeterClient(conn)

// 		// Contact the server and print out its response.
// 		name := defaultName
// 		if len(os.Args) > 1 {
// 			name = os.Args[1]
// 		}
// 		r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
// 		if err != nil {
// 			log.Fatalf("could not greet: %v", err)
// 		}

// 		ret, err := c.GetLastDayAndSecond(context.Background(), &pb.Empty{})
// 		log.Printf("Greeting: %s", r.Message)
// 		log.Println(ret, err)
// }
