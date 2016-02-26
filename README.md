Region à gogo is a microservice, simply returns the country and states/region for a given location.

It uses S2 and a segment tree to create a fast geo shape database, details of implementation are described in this [blog post](http://blog.nobugware.com/post/2016/geo_db_s2_region_polygon).

It can also be used directly from docker `docker run -P akhenakh/regionagogo`

## Data
It uses data from [Natural Earth Data](http://www.naturalearthdata.com/) and performs real points inside tests against geo shapes.

Some regions are not precise enough and some accentuated names are wrong, if you are aware of a better source please tell me.

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