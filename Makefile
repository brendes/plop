APP = plop
LDFLAGS = "-s -w"

.PHONY: build run install clean release

build:
	@ go build -o $(APP) -ldflags=$(LDFLAGS) main.go

run:
	@ go run main.go

install:
	@ go install

clean:
	@ rm -f $(APP) $(APP)-*

release:
	@ GOOS=linux go build -o $(APP)-linux -ldflags=$(LDFLAGS) main.go
	@ GOOS=darwin go build -o $(APP)-darwin -ldflags=$(LDFLAGS) main.go
	@ GOOS=windows go build -o $(APP)-windows -ldflags=$(LDFLAGS) main.go
