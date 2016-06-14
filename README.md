Region à gogo is a microservice, simply returns the country and states/region for a given location.

It uses S2 and a segment tree to create a fast geo shape database, details of implementation are described in this [blog post](http://blog.nobugware.com/post/2016/geo_db_s2_region_polygon).

It can also be used directly from docker `docker run -P akhenakh/regionagogo`

## Data
You can use any geo data but default GeoJSON comes from [Natural Earth Data](http://www.naturalearthdata.com/).  
Some regions are not precise enough and some accentuated names are wrong, if you are aware of a better source please tell me.

Regionagogo is using a hackish but correct region coverer, the officiel Go S2 implementation does not yet use the shape boundaries to perform a region coverage but a rect boundaries.
Regionagogo is reading a file named geodata in msgpack format so you can generate the datafile with another S2 implementation, you can use my C++/ObjC gluecode: [regionagogogen](https://github.com/akhenakh/regionagogogen).  

The image submitted to Docker hub, will always be an optimized one using the C++ implementation.

## Build & Install
```
go get github.com/jteeuwen/go-bindata/...
go get github.com/akhenakh/regionagogo
cd $GOPATH/src/github.com/akhenakh/regionagogo
make
go install github.com/akhenakh/regionagogo/cmd/regionagogo
```

The binary `regionagogo` embed the geodata so it can be copied without any other files.

## Usage
Run `regionagogo`, it will listen on port `8082`.

You can query via HTTP GET:

```
GET /country?lat=19.542915&lng=-155.665857

{
    "code": "US",
    "name": "Hawaii"
}

```

## Using it as a library
You can use it in your own code without the HTTP interface:  

```
gs := regionagogo.NewGeoSearch()
b, _ := ioutil.ReadFile("geodata")
gs.ImportGeoData(b)
r := gs.Query(msg.Latitude, msg.Longitude)
```
