build:
	./rsrc -manifest ddmqtt.exe.manifest
	GOOS=windows GOARCH=amd64 go build .
	mkdir -p /mnt/d/Downloads/ddmqtt/
	cp ddmqtt.exe /mnt/d/Downloads/ddmqtt/
	cp config.json /mnt/d/Downloads/ddmqtt/
