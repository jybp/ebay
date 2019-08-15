.PHONY: test
test:
	go test -count=1 -v -run ./...

.PHONY: integration
integration:
	go test -tags integration -count=1 -v -run "GetItemByLegacyID" ./...