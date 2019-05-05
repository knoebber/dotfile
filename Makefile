cli:
	make -C cli binary
	cp cli/bin/* ./bin/

.PHONY: cli
