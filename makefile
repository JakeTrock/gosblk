test:
	@echo "Running tests"
	@chmod +x ./fake_diskutil
	@alias diskutil=./fake_diskutil
	@diskutil list
	@PATH=".:$$PATH" go test -v ./

devrun:
	@echo "Running lsblk in development mode"
	@go run lsblk.go
build:
	@echo "Building lsblk for ARM and x86"
	@GOOS=darwin GOARCH=arm64 go build -o ./bin/lsblk-arm64 lsblk.go
	@GOOS=darwin GOARCH=amd64 go build -o ./bin/lsblk-amd64 lsblk.go
	@lipo -create -output ./bin/lsblk ./bin/lsblk-arm64 ./bin/lsblk-amd64
	@rm ./bin/lsblk-arm64 ./bin/lsblk-amd64
