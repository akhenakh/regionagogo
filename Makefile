all : build

build : test
	 go build ./...
	
builddocker :
	mkdir -p bindata
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o regionagogo.linux ./cmd/regionagogo 
	docker build -t akhenakh/regionagogo  -f  ./Dockerfile  .

test :
	go test -v  $(glide novendor) 

buildgendb :
	mkdir -p bin
	go build -o bin/ragogenfromjson   ./cmd/ragogenfromjson

generategeodata : clean buildgendb
	./bin/ragogenfromjson  -filename testdata/world_states_10m.geojson -fields iso_a2,name -dbpath ./regiondb -debug

protos :
	protoc -I. geostore.proto --go_out=.

clean :
	rm -f regiondb
	rm -f cmd/regionagogo/regionagogo
	rm -f cmd/ragogenfromjson/ragogenfromjson
	rm -f regionagogo.linux

