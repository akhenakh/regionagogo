import json

f = open("files.txt", 'r')
for line in f:
    json_data = open(line.strip())
    data = json.load(json_data)
    print(data)
