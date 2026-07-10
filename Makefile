default:
	go build -o greyvar-server

protoc:
	rm -rf protocol
	gh repo clone greyvar/protocol
	$(MAKE) -w -C protocol generate
	cp -r protocol/server_go/gen ./

certs: cert-ca cert-server

cert-ca:
	openssl genrsa -out ca.key 4096
	openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.crt -subj "/C=US/ST=CA/L=San Francisco/O=Greyvar/OU=Greyvar/CN=jwr"

cert-server:
	openssl genrsa -out greyvar.key 4096
	openssl req -new -key greyvar.key -out greyvar.csr -config csr.conf
	openssl x509 -req -in greyvar.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out greyvar.crt -days 365 -sha256 -extfile csr.conf -extensions req_ext
	cat ca.crt >> greyvar.crt

.PHONY: default protoc certs
