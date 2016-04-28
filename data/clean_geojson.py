import json

import json
json_data = open('world_states_10m.geojson')
data = json.load(json_data)

features = []
for feature in data["features"]:
    prop = feature["properties"]
    d = {}
    if prop["name"]:
        d["name"] = prop["name"]
    if prop["iso_a2"]:
        d["iso_a2"] = prop["iso_a2"]
    if prop["region"]:
        d["region"] = prop["region"]
    if "name" not in prop:
        d["name"] = prop["region"]

    features.append({"type": "feature", "geometry": feature["geometry"], "properties": d})
data["features"] = features

print json.dumps(data, separators=(',',':'))
