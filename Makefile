build: autogen
	go build ./...

autogen: dep-eventer
	go generate ./...

.PHONY: dep-eventer
dep-eventer:
ifeq (, $(shell which eventer))
	go install github.com/gopherd/tools/cmd/eventer@latest
endif