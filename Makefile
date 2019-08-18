.PHONY: test
test:
	go test -count=1 -v -run ./...

.PHONY: integration
integration:
	go test -count=1 -v ./test/integration -integration=true