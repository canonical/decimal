.PHONY: lint-gosec pre-commit

lint-gosec:
	go run github.com/securego/gosec/v2/cmd/gosec@v2.20.0 ./...

pre-commit:
	pre-commit run --all-files
