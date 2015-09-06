all: lex.yy.c lex.tab.c
.PHONY: all

lex.yy.c: lex.l
	flex -o lex.yy.c lex.l 

lex.tab.h lex.tab.c: lex.y
	bison -d lex.y -o lex.tab.c

clean:
	rm -rf lex.yy.c lex.tab.c lex.tab.h cprog _obj

cprog: lex.yy.cpp lex.tab.cpp
	g++ lex.yy.cpp lex.tab.cpp -lfl -o cprog

run: cprog
	echo "this is a string" | ./cprog

cgo: lex.go
	go tool cgo lex.go
.PHONY: cgo
