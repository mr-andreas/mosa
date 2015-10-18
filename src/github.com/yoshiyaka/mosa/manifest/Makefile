all: lex.yy.c parser.tab.c
.PHONY: all

lex.yy.c: lex.l
	flex -o lex.yy.c lex.l 

parser.tab.h parser.tab.c: parser.y
	bison -d parser.y -o parser.tab.c

clean:
	rm -rf lex.yy.c parser.tab.c parser.tab.h cprog _obj

cgo: lex.go
	go tool cgo lex.go
.PHONY: cgo
