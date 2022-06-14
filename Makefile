testCoverage:
	go test ./... -test.coverprofile test_result.txt
build:
	go build -o bin/gil main.go
run:
	bin/gil -h