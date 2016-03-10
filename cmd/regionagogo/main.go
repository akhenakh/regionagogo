package main

import (
	"encoding/json"
	"log"
	"strconv"

	"net"
	"net/http"

	"github.com/akhenakh/regionagogo"
	pb "github.com/akhenakh/regionagogo/regionagogoservice"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRegionAGogoServer(grpcServer, s)
	grpcServer.Serve(lis)
}
