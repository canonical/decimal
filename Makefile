.PHONY: lint-gosec pre-commit

lint-gosec:
	go run github.com/securego/gosec/v2/cmd/gosec@latest ./...

pre-commit:
	pre-commit run --all-files
