default:
	go build -o greyvar-server


certs:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout greyvar.key -out greyvar.crt -days 365
