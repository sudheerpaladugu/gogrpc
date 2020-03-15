package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gogrpc/greet/greetpb"

	"google.golang.org/grpc"
)

type server struct{}

// registering Greet service with the server
func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet service function was invoked with %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName

	res := &greetpb.GreetResponse{
		Result: result,
	}

	return res, nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Printf("Long running Greet service function was invoked \n")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			//we have finished reading the stream request
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream %v:", req)
		}
		firstName := req.GetGreeting().GetFirstName()
		result += "Hello " + firstName + "! "
	}
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, streem greetpb.GreetService_GreetManyTimesServer) error {

	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		streem.Send(res)

		time.Sleep(1000 * time.Microsecond)
	}
	return nil
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {

	fmt.Printf("GreetEveryone function was invoked with a streaming request \n")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v\n", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "! "

		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if sendErr != nil {
			log.Fatalf("Error while sending data to client: %v\n", sendErr)
			return sendErr
		}
	}
}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Printf("GreetWithDeadline service function was invoked with %v\n", req)
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			//The client has canceled the request
			fmt.Println("The client canceled the request!")
			return nil, status.Error(codes.Canceled, "the client canceled the request")
		}
		time.Sleep(1 * time.Second)
	}

	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName

	res := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}

	return res, nil
}

// registering Greet service with the server
func main() {
	fmt.Println("Hello world!")

	//lis, err := net.Listen("tcp", "0.0.0.0:50051")
	lis, err := net.Listen("tcp", "localhost:50051")

	if err != nil {
		log.Fatalf("Failed to listen : %v", err)
	}
	certFile := "ssl/server.crt"
	keyFile := "ssl/server.pem"
	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)

	if sslErr != nil {
		log.Fatalf("Failed loading certificates: %v", sslErr)
		return
	}
	opts := grpc.Creds(creds)

	s := grpc.NewServer(opts)
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
