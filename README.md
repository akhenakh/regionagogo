Regionagogo is a microservice, simply returning the country for a given location.

It uses S2 and a segment tree to create a fast geo shape database, details of implementation are described in this [blog post](http://blog.nobugware.com/post/2016/geo_db_s2_region_polygon).



It can also be used directly from docker `docker run -P akhenakh/regionagogo`

## Build & Install
```
make
go install github.com/akhenakh/regionagogo/cmd/regionagogo
```

The binary `regionagogo` embed the geodata so it can be copied without any other files.

## Usage
Run `regionagogo`, it will listen on port `8082`.

You can query via HTTP GET:

```
GET /country?lat=48.864716&lng=2.349014

{ "Country": "FR" }
```