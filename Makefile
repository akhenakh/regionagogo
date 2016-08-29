all : build

build : test
	 go build ./...
	
builddocker :
	mkdir -p bindata
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o regionagogo.linux ./cmd/regionagogo 
	docker build -t akhenakh/regionagogo  -f  ./Dockerfile  .

test :
	go test -v ./...

buildgendata :
	mkdir -p bin
	go build -o bin/gendata  ./cmd/gendata 

generategeodata : clean buildgendata
	./bin/gendata  -filename testdata/world_states_10m.geojson -fields iso_a2,name -dbpath ./regiondb -debug

protos :
	protoc -I. geostore.proto --go_out=.

clean :
	rm -f regiondb
	rm -f cmd/regionagogo/regionagogo
	rm -f cmd/gendata/geodata
	rm -f regionagogo.linux

