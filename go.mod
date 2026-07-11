module github.com/greyvar/server

require (
	github.com/coder/websocket v1.8.12
	github.com/fsnotify/fsnotify v1.8.0
	github.com/greyvar/datlib v0.0.0-20240817083143-af99b6da5ab9
	github.com/jamesread/golure v0.0.0-20250821142658-fa4de6860090
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/greyvar/datlib => ../datlib

require golang.org/x/sys v0.28.0 // indirect

go 1.25.0
