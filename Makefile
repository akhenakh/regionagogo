all: build

build: test 
	 go build ./...
	
builddocker:
	mkdir -p bindata
	go-bindata -nomemcopy ./bindata
	mv bindata.go cmd/regionagogo
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o regionagogo.linux ./cmd/regionagogo 
	docker build -t akhenakh/regionagogo  -f  ./Dockerfile  .

test: generatebindata 
	go test -v ./...

gendata:
	mkdir -p bin
	go build -o bin/gendata  ./cmd/gendata 

generategeodata: gendata
	mkdir -p bindata
	./bin/gendata  -filename data/world_states_10m.geojson -fields iso_a2,name
	mv geodata bindata 

generatebindata: generategeodata 
	go-bindata -nomemcopy ./bindata
	mv bindata.go cmd/regionagogo

clean:
	rm -fr bin
	rm -fr bindata
	rm -f cmd/regionagogo/bindata.go
	rm -f cmd/regionagogo/regionagogo
	rm -f cmd/gendata/geodata
	rm -f regionagogo.linux

