examples: bin/mboxserver bin/mboximapclient
	file bin/*
help:
	@echo "make [examples|test|clean|distclean]"
test:
	go test -v ./...
bin/mboxserver: examples/websrv/*.go *.go 
	go build -o $@ ./examples/websrv
bin/mboximapclient: examples/imap/*.go *.go 
	go build -o $@ ./examples/imap
clean:
	${RM} -r bin
distclean: clean
	${RM} *.mbox