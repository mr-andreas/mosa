%{
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "_cgo_export.h"


// stuff from flex that bison needs to know about:
//extern "C" int yylex();
//extern "C" int yyparse();
//extern "C" int doparse();
extern int line_num;
extern FILE *yyin;

typedef struct yy_buffer_state * YY_BUFFER_STATE;
extern int yyparse();
extern YY_BUFFER_STATE yy_scan_string(char * str);
extern void yy_delete_buffer(YY_BUFFER_STATE buffer);

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
		//cout << "saw class " << $2 << endl;
		$$ = NewClass($2, $4);
	}

defs:
	defs ',' def { AddDefs($1, $3);  }
	| def        { $$ = NewDefs($1); }
	|		     { $$ = NilDefs();   }
	;
	
def:
	STRING '=' STRING {
		//cout << "saw def " << $1 << " = " << $3 << endl;
		$$ = SawDef($1, $3);
	}

%%

char *last_error = NULL;

t_error doparse(char *file) {
	int ret;
	
	YY_BUFFER_STATE buffer = yy_scan_string(file);
    ret = yyparse();
    yy_delete_buffer(buffer);
	
	t_error err;
	err.code = ret;
	err.error = last_error;
	err.line = line_num;
	
	return err;
}

void yyerror(const char *s) {
	if(!last_error) {
		free(last_error);
	}
	last_error = strdup(s);
}
