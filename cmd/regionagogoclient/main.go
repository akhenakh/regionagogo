package main

import (
	"context"
	"flag"
	"log"

	pb "github.com/akhenakh/regionagogo/regionagogosvc"
	"google.golang.org/grpc"
)

var (
	grpcURL = flag.String("gprcURL", "127.0.0.1:8083", "The address of the gRPC server")
	lat     = flag.Float64("lat", 48.8, "Latitude to query")
	lng     = flag.Float64("lng", 2.2, "Longitude to query")
)

func main() {

	flag.Parse()

	conn, err := grpc.Dial(*grpcURL, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRegionAGogoClient(conn)

	r, err := client.GetRegion(context.Background(), &pb.Point{Latitude: float32(*lat), Longitude: float32(*lng)})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(r)

}
