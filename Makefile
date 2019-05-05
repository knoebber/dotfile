.PHONY: cli
cli:
	make -C cli binary
	cp cli/bin/dotfile ./bin/
