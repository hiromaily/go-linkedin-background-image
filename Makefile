bld:
	go build -i -v -o ${GOPATH}/bin/goimage main.go

run:
	go run main.go -j $(PWD)/jsons/preference.json

genjson:
	gogentype -file $(PWD)/jsons/preference.json