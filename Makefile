test:
	make -C cli test

ci_test:
	make -C cli ci_test
cli:
	make -C cli binary
	cp cli/bin/dotfile ./bin/
.PHONY: cli
