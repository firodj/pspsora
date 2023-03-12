TARGET=pspsora

run: build
	./${TARGET}

build:
	go build -v -o ${TARGET}