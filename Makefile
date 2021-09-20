default: help
## This help screen.
help:
	@printf "Available targets:\n\n"
	@awk '/^[a-zA-Z\-\_0-9%:\\]+/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = $$1; \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub("\\\\", "", helpCommand); \
			gsub(":+$$", "", helpCommand); \
			printf "  \x1b[32;01m%-35s\x1b[0m %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST) | sort -u
	@printf "\n"

## Runs unit tests
test:
	@go test ./... -v -tags=test

## Runs unit tests with coverage
test-coverage:
	@go test ./... -coverprofile=coverage.out -tags=test

## builds snapshot with goreleaser
build:
	@goreleaser --snapshot --skip-validate --skip-publish --rm-dist