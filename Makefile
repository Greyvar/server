default:
	go build -o greyvar-server

protoc:
	gh repo clone greyvar/protocol
	$(MAKE) -w -C protocol generate
	cp -r protocol/server_go/gen ./


certs:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout greyvar.key -out greyvar.crt -days 365

.PHONY: default protoc certs
