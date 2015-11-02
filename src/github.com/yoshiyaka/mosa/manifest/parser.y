%{
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "_cgo_export.h"
#include "types.h"

extern int line_num;
extern int level;
extern FILE *yyin;
extern int yylineno;

typedef struct yy_buffer_state * YY_BUFFER_STATE;
extern int yyparse();
extern YY_BUFFER_STATE yy_scan_string(char * str);
extern void yy_delete_buffer(YY_BUFFER_STATE buffer);

int yylex();
void yylex_destroy();
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
%token DEFINE
%token NODE
%token ARROW
%token <sval> QUOTED_STRING
%token <sval> INTPOL_TEXT
%token <sval> INTPOL_VARIABLE

%type <gohandle> def
%type <gohandle> defs
%type <gohandle> define
%type <gohandle> define_body
%type <gohandle> class
%type <gohandle> node
%type <gohandle> optional_arg_defs
%type <gohandle> define_arg_defs
%type <gohandle> arg_defs
%type <gohandle> arg_def
%type <gohandle> file_body
%type <gohandle> file
%type <gohandle> declaration
%type <gohandle> variable_def
%type <gohandle> proplist
%type <gohandle> prop
%type <gohandle> value
%type <gohandle> interpolated_string
%type <gohandle> interpolated_string_value
%type <gohandle> arrayentries
%type <gohandle> array
%type <gohandle> scalar
%type <gohandle> string_or_var
%type <gohandle> reference

%%

file:
	file_body				{ NewFile($1); }
	| /* Empty manifest */	{ NewFile(NilArray(ASTTYPE_ARRAY_INTERFACE)); }

file_body:
	  file_body class   	{ $$ = AppendArray($1, $2); }
	| file_body define		{ $$ = AppendArray($1, $2); }
	| file_body node		{ $$ = AppendArray($1, $2); }
	| class					{ $$ = AppendArray(NilArray(ASTTYPE_ARRAY_INTERFACE), $1); }
	| define				{ $$ = AppendArray(NilArray(ASTTYPE_ARRAY_INTERFACE), $1); }
	| node					{ $$ = AppendArray(NilArray(ASTTYPE_ARRAY_INTERFACE), $1); }

node:
	  NODE QUOTED_STRING '{' defs '}' { $$ = SawNode(@1.first_line, $2, $4); }
	| NODE QUOTED_STRING '{' '}' { $$ = SawNode(@1.first_line, $2, NilArray(ASTTYPE_DEFS)); }

define:
	DEFINE STRING STRING define_arg_defs define_body {
		$$ = SawDefine(@1.first_line, $2, $3, $4, $5);
		if($$ == -1) {
			yyerror("Expected 'single' or 'multiple' after define");
			YYABORT;
		}
	}

define_arg_defs:
	  '(' ')'			{ $$ = NilArray(ASTTYPE_ARGDEFS); }
	| '(' arg_defs ')'	{ $$ = $2; }

define_body:
	  '{' '}'		{ $$ = NilArray(ASTTYPE_DEFS); }
	| '{' defs '}'	{ $$ = $2; }

class:
	  CLASS STRING optional_arg_defs '{' defs '}'	{ $$ = NewClass(@1.first_line, $2, $3, $5);						}
	| CLASS STRING optional_arg_defs '{' '}'		{ $$ = NewClass(@1.first_line, $2, $3, NilArray(ASTTYPE_DEFS));	}

optional_arg_defs:
	/* empty */					{ $$ = NilArray(ASTTYPE_ARGDEFS); }
	| '(' ')'					{ $$ = NilArray(ASTTYPE_ARGDEFS); }
	| '(' arg_defs ')'			{ $$ = $2; }

arg_defs:
	  arg_defs arg_def			{ $$ = AppendArray($1, $2); }
	| arg_def					{ $$ = AppendArray(NilArray(ASTTYPE_ARGDEFS), $1); }

arg_def:
	  VARIABLENAME ','				{ $$ = SawArgDef(@1.first_line, $1, 0);  }
	| VARIABLENAME '=' scalar ','	{ $$ = SawArgDef(@1.first_line, $1, $3); }
	| VARIABLENAME '=' array  ','	{ $$ = SawArgDef(@1.first_line, $1, $3); }

defs:
	  defs def	{ $$ = AppendArray($1, $2);						}
	| def		{ $$ = AppendArray(NilArray(ASTTYPE_DEFS), $1);	}
	
def:
	variable_def | declaration;
	
variable_def:
	VARIABLENAME '=' value { $$ = SawVariableDef(@1.first_line, $1, $3);	}

declaration:
	  STRING '{' string_or_var ':' proplist '}'	{ $$ = SawDeclaration(@1.first_line, $1, $3, $5); }
	| STRING '{' string_or_var ':' '}'			{ $$ = SawDeclaration(@1.first_line, $1, $3, NilArray(ASTTYPE_PROPLIST)); }

proplist:
	  proplist prop	{ $$ = AppendArray($1, $2); }
	| prop			{ $$ = AppendArray(NilArray(ASTTYPE_PROPLIST), $1); }
	;

prop:
	STRING ARROW value ','	{ $$ = SawProp(@1.first_line, $1, $3); }

value:
	  scalar		{ $$ = $1; }
	| array			{ $$ = $1; }
	| reference		{ $$ = $1; }

scalar:
	  QUOTED_STRING			{ $$ = SawQuotedString(@1.first_line, $1);			}
	| interpolated_string	{ $$ = $1;											}
	| VARIABLENAME			{ $$ = SawVariableName(@1.first_line, $1);			}
	| INT					{ $$ = SawInt(@1.first_line, $1);					}

string_or_var:
	  QUOTED_STRING			{ $$ = SawQuotedString(@1.first_line, $1);			}
	| interpolated_string	{ $$ = $1;											}
	| VARIABLENAME			{ $$ = SawVariableName(@1.first_line, $1);			}

reference:
	STRING '[' scalar ']' { $$ = SawReference(@1.first_line, $1, $3); }

array:
	  '[' arrayentries ']'	{ $$ = $2; }
	| '[' ']' 				{ $$ = NilArray(ASTTYPE_ARRAY); }

arrayentries:
	  arrayentries value ','	{ $$ = AppendArray($1, $2); }
	| value ','					{ $$ = AppendArray(NilArray(ASTTYPE_ARRAY), $1); }

interpolated_string:
	  interpolated_string interpolated_string_value	{ $$ = AppendInterpolatedString($1, $2);		}
	| interpolated_string_value						{ $$ = AppendInterpolatedString(EmptyInterpolatedString(@1.first_line), $1);	}

interpolated_string_value:
	  INTPOL_VARIABLE	{ $$ = SawVariableName(@1.first_line, $1); }
	| INTPOL_TEXT 		{ $$ = SawString($1); }
	
%%

char *last_error = NULL;

t_error doparse(char *file) {
	int ret;
	line_num = 1;
	level = 0;
	
	memset(&yylloc, 0, sizeof(YYLTYPE));
	yylineno = 1;
	
	YY_BUFFER_STATE buffer = yy_scan_string(file);
    ret = yyparse();
    yy_delete_buffer(buffer);
	yylex_destroy();
	
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
