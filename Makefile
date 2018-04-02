all:
	go build -o ww main.go

install:
	mv ww ${GOPATH}/bin