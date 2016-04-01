package main

import (
	"encoding/json"
	"flag"
	"log"
	"strconv"

	"net"
	"net/http"

	"github.com/akhenakh/regionagogo"
	pb "github.com/akhenakh/regionagogo/regionagogoservice"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "testdata/server1.pem", "The TLS cert file")
	keyFile  = flag.String("key_file", "testdata/server1.key", "The TLS key file")
)

type Server struct {
	*regionagogo.GeoSearch
}

func (s *Server) GetRegion(ctx context.Context, p *pb.Point) (*pb.RegionResponse, error) {
	region := s.Query(float64(p.Latitude), float64(p.Longitude))
	if region == nil {
		return &pb.RegionResponse{Code: "unknown", Name: "unknown"}, nil
	}

	rs := pb.RegionResponse{Code: region.Code, Name: region.Name}
	return &rs, nil
}

// countryHandler takes a lat & lng query params and return a JSON
// with the country of the coordinate
func (s *Server) countryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	slat := query.Get("lat")
	lat, err := strconv.ParseFloat(slat, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	slng := query.Get("lng")
	lng, err := strconv.ParseFloat(slng, 64)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	region := s.Query(lat, lng)
	w.Header().Set("Content-Type", "application/json")

	if region == nil {
		js, _ := json.Marshal(regionagogo.Region{Code: "unknown", Name: "unknown"})
		w.Write(js)
		return
	}

	js, _ := json.Marshal(region)
	w.Write(js)
}

func main() {
	flag.Parse()

	gs := regionagogo.NewGeoSearch()

	data, err := Asset("bindata/geodata")
	if err != nil {
		log.Fatal(err)
	}

	err = gs.ImportGeoData(data)
	if err != nil {
		log.Fatal(err)
	}
	s := &Server{GeoSearch: gs}

	http.HandleFunc("/country", s.countryHandler)
	go func() { http.ListenAndServe(":8082", nil) }()

	lis, err := net.Listen("tcp", ":8083")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			grpclog.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRegionAGogoServer(grpcServer, s)
	grpcServer.Serve(lis)
}
