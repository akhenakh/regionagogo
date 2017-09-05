#!/bin/sh

wget https://whosonfirst.mapzen.com/bundles/wof-region-latest-bundle.tar.bz2
tar jxvf wof-region-latest-bundle.tar.bz2
find ~/GIS/wofregion -type f -name "*.geojson" -ls -exec ../cmd/ragogenfromjson/ragogenfromjson -dbpath mygeodb  -filename {} -importFields "wof:country" -renameFields "wof:country=iso" -featureImport \;

