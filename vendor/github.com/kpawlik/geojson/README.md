#GEOJSON
***

Go package to easy and quick create datastructure which can be serialized to geojson format

***

## INSTALLATION

    $ go get https://github.com/kpawlik/geojson
    $ go install https://github.com/kpawlik/geojson

***

## USAGE EXAMPLE

```go
package main

import (
    "fmt"
    gj "github.com/kpawlik/geojson"
)

func main() {
    fc := gj.NewFeatureCollection([]*gj.Feature {})

    // feature
    p := gj.NewPoint(gj.Coordinate{12, 3.123})
    f1 := gj.NewFeature(p, nil, nil)
    fc.AddFeatures(f1)

    // feature with propertises
    props := map[string]interface{}{"name": "location", "code": 107}
    f2 := gj.NewFeature(p, props, nil)
    fc.AddFeatures(f2)

    // feature with propertises and id
    f3 := gj.NewFeature(p, props, 11101)
    fc.AddFeatures(f3)

    ls := gj.NewLineString(gj.Coordinates{{1, 1}, {2.001, 3}, {4001, 1223}})
    f4 := gj.NewFeature(ls, nil, nil)
    fc.AddFeatures(f4)

    if gjstr, err := gj.Marshal(fc); err != nil {
        panic(err)
    } else {
        fmt.Println(gjstr)
    }
}
```
