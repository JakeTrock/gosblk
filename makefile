test:
	@echo "Running tests"
	@chmod +x ./fake_diskutil
	@alias diskutil=./fake_diskutil
	@diskutil list
	@PATH=".:$$PATH" go test -v ./

devrun:
	@echo "Running lsblk in development mode"
	@go run main.go
build:
	@echo "Building lsblk"
	@go build -o ./bin/lsblk main.go
