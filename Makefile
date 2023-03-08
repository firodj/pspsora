TARGET=pspsora

run:
	go build -o ${TARGET} && ./${TARGET}
