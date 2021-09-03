build:
	cd utils/grpc/proto && protoc --go_out=plugins=grpc:. *.proto && cd ../../../ && env GOOS=$(OS) go build -o asterisk-ari

run:
	./asterisk-ari
