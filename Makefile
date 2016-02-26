all: build

build: test 
	 go build ./...

test: generatebindata 
	go test -v ./...

gendata:
	mkdir -p bin
	go build -o bin/gendata  ./cmd/gendata 

generategeodata: gendata
	mkdir -p bindata
	./bin/gendata  -filename data/world_states_10m.geojson 
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

