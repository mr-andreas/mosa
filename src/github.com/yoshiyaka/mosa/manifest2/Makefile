all: lex.yy.cpp lex.tab.cpp
.PHONY: all

lex.yy.cpp: lex.l
	flex -o lex.yy.cpp lex.l 

lex.tab.hpp: lex.tab.cpp
lex.tab.cpp: lex.y
	bison -d lex.y -o lex.tab.cpp

clean:
	rm -f lex.yy.cpp lex.tab.cpp lex.tab.hpp cprog

cprog: lex.yy.cpp lex.tab.cpp
	g++ lex.yy.cpp lex.tab.cpp -lfl -o cprog

run: cprog
	echo "this is a string" | ./cprog
