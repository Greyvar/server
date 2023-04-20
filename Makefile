default:
	go build -o greyvar-server github.com/greyvar/server/cmd/servercmd/ 


certs:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout greyvar.key -out greyvar.crt -days 365
