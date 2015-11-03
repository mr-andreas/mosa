LEX_DIR = src/github.com/yoshiyaka/mosa/manifest
# BIN = src/github.com/yoshiyaka/mosa/mosa/mosa

bin/mosa: $(wildcard *go) lex
	GOPATH=`pwd` go build github.com/yoshiyaka/mosa/...
	GOPATH=`pwd` go install github.com/yoshiyaka/mosa/...

lex:
	$(MAKE) -C $(LEX_DIR)
.PHONY: lex

clean:
	rm -f $(BIN)
	$(MAKE) -C $(LEX_DIR) clean
.PHONY: clean

run: bin/mosa
	bin/mosa src/github.com/yoshiyaka/mosa/testdata

test: lex
	GOPATH=`pwd` go test github.com/yoshiyaka/mosa/...
.PHONY: test