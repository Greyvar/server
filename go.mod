module github.com/greyvar/server

replace github.com/greyvar/datlib => ../datlib/

require (
	github.com/fsnotify/fsnotify v1.5.4
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/greyvar/datlib v0.0.0-20220723191212-08d2466064bf
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.7.1 // indirect
	golang.org/x/sys v0.7.0 // indirect
	google.golang.org/protobuf v1.28.0
	nhooyr.io/websocket v1.8.7
)

go 1.13
