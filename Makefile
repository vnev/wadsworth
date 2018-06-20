all: main.go
	go build -o ww main.go
	@$(MAKE) install

install: ww
	mv ww ${GOPATH}/bin
