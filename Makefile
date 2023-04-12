TARGET=pspsora

run: build
	./${TARGET}

build:
	go build -v -o ${TARGET}

test:
	go test ./...

cover:
	go test -coverprofile cover.out ./...
	go tool cover -func=cover.out
