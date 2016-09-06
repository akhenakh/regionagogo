[![wercker status](https://app.wercker.com/status/82617ea8fffe1b93a6956c7ba5559365/s/master "wercker status")](https://app.wercker.com/project/byKey/82617ea8fffe1b93a6956c7ba5559365) [![Go Report Card](https://goreportcard.com/badge/github.com/akhenakh/regionagogo)](https://goreportcard.com/report/github.com/akhenakh/regionagogo)

Region à gogo is a microservice, it's a simple database that returns metadata associated to a fence for a given location.

It uses S2 and a segment tree to create a fast geo shape database, details of implementation are described in this [blog post](http://blog.nobugware.com/post/2016/geo_db_s2_region_polygon).

It can also be used directly from docker `docker run -P akhenakh/regionagogo`

## Data
You can use any geo data but the provided GeoJSON comes from [Natural Earth Data](http://www.naturalearthdata.com/).
Some regions are not precise enough and some accentuated names are wrong, if you are aware of a better source please tell me.

It works too with the better [Gadm Data](http://gadm.org/version2) but the data are not free for commercial use.    

Regionagogo is using a BoltDB datafile to store the fences but the small segment tree lives in memory.  

## Build & Install
```
go get github.com/akhenakh/regionagogo
cd $GOPATH/src/github.com/akhenakh/regionagogo
make
```

To generate the database from GeoJSON use the provided `ragogenfromjson` command, you can specify the fields you want from the GeoJSON properties to be saved into the DB:
```
ragogenfromjson -filename testdata/world_states_10m.geojson -fields iso_a2,name -dbpath ./region.db
```

## Usage
Run `regionagogo -dbpath ./region.db`, it will listen on port `8082`.

You can query via HTTP GET:

```
GET /query?lat=19.542915&lng=-155.665857

{
    "code": "US",
    "name": "Hawaii"
}

```

## Using it as a library
You can use it in your own code without the HTTP interface:  

```
gs := regionagogo.NewGeoSearch("region.db")
gs.ImportGeoData()
r := gs.StabbingQuery(msg.Latitude, msg.Longitude)
```
