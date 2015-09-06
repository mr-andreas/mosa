%{
#include <cstdio>
#include <iostream>
using namespace std;

extern "C" {
#include "_cgo_export.h"
}

// stuff from flex that bison needs to know about:
//extern "C" int yylex();
//extern "C" int yyparse();
//extern "C" int doparse();
extern "C" int line_num;
extern "C" FILE *yyin;

//extern "C" void GoFunc();

GoInterface x;
 
void yyerror(const char *s);
%}

// Bison fundamentally works by asking flex to get the next token, which it
// returns as an object of type "yystype".  But tokens could be of any
// arbitrary data type!  So we deal with that in Bison by defining a C union
// holding each of the types of tokens that Flex could return, and have Bison
// use that union instead of "int" for the definition of "yystype":
%union {
  int ival;
  float fval;
  char *sval;
  int gohandle;
}

%token SNAZZLE TYPE
%token END

// define the "terminal symbol" token types I'm going to use (in CAPS
// by convention), and associate each with a field of the union:
%token <ival> INT
%token <fval> FLOAT
%token <sval> STRING
%token <sval> CLASS

%type <gohandle> def
%type <gohandle> defs
%type <gohandle> class
%type <gohandle> classes
%type <gohandle> file

%%

// the first rule defined is the highest-level rule, which in our
// case is just the concept of a whole "snazzle file":
/*snazzle:
  header template body_section footer { cout << "done with a snazzle file!" << endl; }
  ;
header:
  SNAZZLE FLOAT { cout << "reading a snazzle file version " << $2 << endl; }
  ;
template:
  typelines
  ;
typelines:
  typelines typeline
  | typeline
  ;
typeline:
  TYPE STRING { cout << "new defined snazzle type: " << $2 << endl; }
  ;
body_section:
  body_lines
  ;
body_lines:
  body_lines body_line
  | body_line
  ;
body_line:
  INT INT INT INT STRING { cout << "new snazzle: " << $1 << $2 << $3 << $4 << $5 << endl; }
  ;
footer:
  END
  ;*/

file:
	classes			{ NewFile($1); }
	;

classes:
	classes class   { AddClasses($1, $2); }
	| class			{ $$ = NewClasses($1); }
	;

class:
	CLASS STRING '{' defs '}' {
		cout << "saw class " << $2 << endl;
		$$ = NewClass($2, $4);
	}

defs:
	defs ',' def { AddDefs($1, $3); }
	| def        { $$ = NewDefs($1); }
	;
	
def:
	STRING '=' STRING {
		cout << "saw def " << $1 << " = " << $3 << endl;
		$$ = SawDef($1, $3);
	}

%%

int main2(int, char**) {
  // open a file handle to a particular file:
  FILE *myfile = fopen("a.snazzle.file", "r");
  // make sure it is valid:
  if (!myfile) {
    cout << "I can't open a.snazzle.file!" << endl;
    return -1;
  }
  // set flex to read from it instead of defaulting to STDIN:
  yyin = myfile;
  
  // parse through the input until there is no more:
  do {
    yyparse();
  } while (!feof(yyin));
  
}

int doparse() {
	printf("doparse called\n");
	GoFunc();
	
  do {
    yyparse();
  } while (!feof(yyin));
  
  return 0;
}

void yyerror(const char *s) {
  cout << "EEK, parse error on line " << line_num << "!  Message: " << s << endl;

  // might as well halt now:
  exit(-1);
}
