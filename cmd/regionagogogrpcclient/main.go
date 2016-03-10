package main

import (
	"log"

	pb "github.com/akhenakh/regionagogo/regionagogoservice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {

	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("127.0.0.1:8083", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRegionAGogoClient(conn)

	r, err := client.GetRegion(context.Background(), &pb.Point{40.9146138, -7.46188906})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(r)
}
