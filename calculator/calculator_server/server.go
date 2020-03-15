package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc/reflection"

	"github.com/gogrpc/calculator/calculatorpb"
	"google.golang.org/grpc"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	firstNumber := req.FirstNumber
	secondNumer := req.SecondNumber

	sum := firstNumber + secondNumer
	res := &calculatorpb.SumResponse{
		SumResult: sum,
	}

	return res, nil
}

func (*server) PrimeNumberDecompositioin(req *calculatorpb.PrimeNumberDecompositioinRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositioinServer) error {
	fmt.Printf("Received PrimeNumberDecomsition RPC: %v\n", req)
	inputNumber := req.GetNumber()

	divisor := int64(2)

	for inputNumber > 1 {
		if inputNumber%divisor == 0 {
			stream.Send(&calculatorpb.PrimeNumberDecompositioinResponse{
				PrimeFactor: divisor,
			})

			inputNumber = inputNumber / divisor
		} else {
			divisor++
			fmt.Printf("Divisior has incresead to %v\n", divisor)
		}

	}

	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("Received ComputeAverate stream RPC request received  ...\n")

	sum := int64(0)
	count := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			average := float64(sum) / float64(count)
			fmt.Printf("Returning ComputedAverate :%v\n", average)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: average,
			})
		}

		if err != nil {
			log.Fatalf("Error whiel processing stream request: %v", err)
		}

		sum += req.GetNumber()
		count++
	}

}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("Received FindMaximum stream RPC request received  ...\n")

	maxNumber := int64(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		number := req.GetNumber()

		if number > maxNumber {
			maxNumber = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: maxNumber,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending data to client: %v", err)
				return err
			}
		}
	}

}

func main() {
	fmt.Println("Calculator Server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to Listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
