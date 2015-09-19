%{
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "_cgo_export.h"
#include "types.h"


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
%token <sval> VARIABLE
%token <sval> CLASS
%token <sval> QUOTED_STRING

%type <gohandle> def
%type <gohandle> defs
%type <gohandle> class
%type <gohandle> classes
%type <gohandle> file

%%

file:
	classes			{ NewFile($1); }
	;

classes:
	classes class   { $$ = AppendArray($1, $2); }
	| class			{ $$ = AppendArray(NilArray(ASTTYPE_CLASSES), $1); }
	;

class:
	CLASS STRING '{' defs '}' {
		$$ = NewClass($2, $4);
	}

defs:
	defs ',' def { $$ = AppendArray($1, $3);  }
	| def        { $$ = AppendArray(NilArray(ASTTYPE_DEFS), $1); }
	|		     { $$ = NilArray(ASTTYPE_DEFS);   }
	;
	
def:
	VARIABLE '=' QUOTED_STRING	{ $$ = SawDef($1, $3);	}
	| VARIABLE '=' VARIABLE		{ $$ = SawDef($1, $3);	}

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
