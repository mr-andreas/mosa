%{
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "_cgo_export.h"
#include "types.h"

extern int line_num;
extern FILE *yyin;
extern int yylineno;

typedef struct yy_buffer_state * YY_BUFFER_STATE;
extern int yyparse();
extern YY_BUFFER_STATE yy_scan_string(char * str);
extern void yy_delete_buffer(YY_BUFFER_STATE buffer);

void yyerror(const char *s);

%}

%define parse.error verbose
%locations

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

// define the "terminal symbol" token types I'm going to use (in CAPS
// by convention), and associate each with a field of the union:
%token <ival> INT
%token <fval> FLOAT
%token <sval> STRING
%token <sval> VARIABLENAME
%token CLASS
%token ARROW
%token <sval> QUOTED_STRING

%type <gohandle> def
%type <gohandle> defs
%type <gohandle> class
%type <gohandle> classes
%type <gohandle> file
%type <gohandle> declaration
%type <gohandle> variable_def
%type <gohandle> proplist
%type <gohandle> prop
%type <gohandle> value
%type <gohandle> arrayentries
%type <gohandle> array
%type <gohandle> scalar
%type <gohandle> reference

%%

file:
	classes			{ NewFile($1); }

classes:
	classes class   { $$ = AppendArray($1, $2); }
	| class			{ $$ = AppendArray(NilArray(ASTTYPE_CLASSES), $1); }

class:
	CLASS STRING '{' defs '}'	{ $$ = NewClass(@1.first_line, $2, $4);						}
	| CLASS STRING '{' '}'		{ $$ = NewClass(@1.first_line, $2, NilArray(ASTTYPE_DEFS));	}

defs:
	defs def	{ $$ = AppendArray($1, $2);						}
	| def		{ $$ = AppendArray(NilArray(ASTTYPE_DEFS), $1);	}
	
def:
	variable_def | declaration;
	
variable_def:
	VARIABLENAME '=' value { $$ = SawVariableDef(@1.first_line, $1, $3);	}

declaration:
	STRING '{' scalar ':' proplist '}'	{ $$ = SawDeclaration(@1.first_line, $1, $3, $5); }
	| STRING '{' scalar':' '}'			{ $$ = SawDeclaration(@1.first_line, $1, $3, NilArray(ASTTYPE_PROPLIST)); }

proplist:
	proplist prop	{ $$ = AppendArray($1, $2); }
	| prop			{ $$ = AppendArray(NilArray(ASTTYPE_PROPLIST), $1); }
	;

prop:
	STRING ARROW value ','	{ $$ = SawProp(@1.first_line, $1, $3); }

value:
	scalar			{ $$ = $1; }
	| array			{ $$ = $1; }
	| reference		{ $$ = $1; }

scalar:
	QUOTED_STRING	{ $$ = SawQuotedString(@1.first_line, $1);	}
	| VARIABLENAME	{ $$ = SawVariableName(@1.first_line, $1);		}
	| INT			{ $$ = SawInt(@1.first_line, $1);			}

reference:
	STRING '[' scalar ']' { $$ = SawReference(@1.first_line, $1, $3); }

array:
	'[' arrayentries ']'	{ $$ = $2; }
	| '[' ']' 				{ $$ = NilArray(ASTTYPE_ARRAY); }

arrayentries:
	arrayentries value ','	{ $$ = AppendArray($1, $2); }
	| value ','				{ $$ = AppendArray(NilArray(ASTTYPE_ARRAY), $1); }

%%

char *last_error = NULL;

t_error doparse(char *file) {
	int ret;
	line_num = 0;
	
	memset(&yylloc, 0, sizeof(YYLTYPE));
	yylineno = 1;
	
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
