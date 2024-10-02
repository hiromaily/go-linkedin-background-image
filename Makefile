# .PHONY: gen-json
# gen-json:
# 	gogentype -file $(PWD)/jsons/preference.json

.PHONY: lint
lint:
	go fmt ./...
	go vet ./...

.PHONY: build
build:
	go build -i -v -o ${GOPATH}/bin/goimage main.go

run:
	go run main.go -j $(PWD)/jsons/preference.json
