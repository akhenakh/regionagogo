all : build

build : test
	 go build ./...
	
builddocker :
	mkdir -p bindata
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o regionagogo.linux ./cmd/regionagogo 
	docker build -t akhenakh/regionagogo  -f  ./Dockerfile  .

test : generategeodata
	go test -v ./...

buildgendata :
	mkdir -p bin
	go build -o bin/gendata  ./cmd/gendata 

generategeodata : buildgendata
	./bin/gendata  -filename data/world_states_10m.geojson -fields iso_a2,name -dbpath ./regiondb
	mv geodata bindata

protos :
	protoc -I. geostore.proto --go_out=.

clean :
	rm -fr bin
	rm -fr bindata
	rm -f cmd/regionagogo/regionagogo
	rm -f cmd/gendata/geodata
	rm -f regionagogo.linux

