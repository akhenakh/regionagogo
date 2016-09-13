all : build

build : test
	 go build ./cmd/... 
	 go build ./geostore/... 
	 go build .
	
builddocker : generategeodata
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o regionagogo.linux -a -installsuffix cgo ./cmd/regionagogo 
	docker build -t akhenakh/regionagogo -f ./Dockerfile  .

test :
	go test -v . ./cmd/... ./geostore/... ./db/...

bin/ragogenfromjson :
	mkdir -p bin
	go build -o bin/ragogenfromjson ./cmd/ragogenfromjson

generategeodata : region.db

region.db : bin/ragogenfromjson
	./bin/ragogenfromjson -filename testdata/world_states_10m.geojson -fields iso_a2,name -dbpath $@ 

protos :
	go generate github.com/akhenakh/regionagogo/geostore

clean :
	rm -f region.db
	rm -f cmd/regionagogo/regionagogo
	rm -f cmd/ragogenfromjson/ragogenfromjson
	rm -f bin/ragogenfromjson
	rm -f regionagogo.linux

