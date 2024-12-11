build:
	./rsrc -manifest ddmqtt.exe.manifest
	GOOS=windows GOARCH=amd64 go build  -o ddmqtt.exe cmd/main.go