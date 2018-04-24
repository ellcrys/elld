pb:
	protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf --gogoslick_out=. ./wire/messages.proto
	echo "> target 'pb' done"
test:
	go test ./... -v