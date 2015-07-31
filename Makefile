GO=go
GET=$(GO) get
BUILD=$(GO) build
RUN=$(GO) run

build:
	$(BUILD) -o meuip.io

run:
	$(RUN) meuip.go

install-dependencies:
	$(GET) github.com/go-martini/martini
