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
%token FUNC
%token ARROW
%token <sval> PLUSMINUS // + -
%token <sval> MULDIV // * /
%token <sval> COMPARISON // > < >= <=
%token <sval> QUOTED_STRING
%token INTPOL_START
%token <sval> INTPOL_TEXT
%token <sval> INTPOL_VARIABLE

%left COMPARISON
%left PLUSMINUS
%left MULDIV

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
%type <gohandle> expression
%type <gohandle> interpolated_string
%type <gohandle> interpolated_string_list
%type <gohandle> interpolated_string_value
%type <gohandle> arrayentries
%type <gohandle> array
%type <gohandle> scalar
%type <gohandle> string_or_var
%type <gohandle> reference

%%

file:
	file_body				{ sawBody($1); }
	| /* Empty manifest */	{}

file_body:
	  file_body class   	{ $$ = appendArray($1, $2); }
	| file_body define		{ $$ = appendArray($1, $2); }
	| file_body node		{ $$ = appendArray($1, $2); }
	| class					{ $$ = appendArray(nilArray(ASTTYPE_ARRAY_INTERFACE), $1); }
	| define				{ $$ = appendArray(nilArray(ASTTYPE_ARRAY_INTERFACE), $1); }
	| node					{ $$ = appendArray(nilArray(ASTTYPE_ARRAY_INTERFACE), $1); }

node:
	  NODE QUOTED_STRING '{' defs '}'	{ $$ = sawNode(@1.first_line, $2, $4); }
	| NODE QUOTED_STRING '{' '}' 		{ $$ = sawNode(@1.first_line, $2, nilArray(ASTTYPE_DEFS)); }

define:
	DEFINE STRING STRING define_arg_defs define_body {
		$$ = sawDefine(@1.first_line, $2, $3, $4, $5);
		if($$ == -1) {
			yyerror("Expected 'single' or 'multiple' after define");
			YYABORT;
		}
	}

define_arg_defs:
	  '(' ')'			{ $$ = nilArray(ASTTYPE_ARGDEFS); }
	| '(' arg_defs ')'	{ $$ = $2; }

define_body:
	  '{' '}'		{ $$ = nilArray(ASTTYPE_DEFS); }
	| '{' defs '}'	{ $$ = $2; }

class:
	  CLASS STRING optional_arg_defs '{' defs '}'	{ $$ = newClass(@1.first_line, $2, $3, $5);						}
	| CLASS STRING optional_arg_defs '{' '}'		{ $$ = newClass(@1.first_line, $2, $3, nilArray(ASTTYPE_DEFS));	}

optional_arg_defs:
	/* empty */					{ $$ = nilArray(ASTTYPE_ARGDEFS); }
	| '(' ')'					{ $$ = nilArray(ASTTYPE_ARGDEFS); }
	| '(' arg_defs ')'			{ $$ = $2; }

arg_defs:
	  arg_defs arg_def			{ $$ = appendArray($1, $2); }
	| arg_def					{ $$ = appendArray(nilArray(ASTTYPE_ARGDEFS), $1); }

arg_def:
	  VARIABLENAME ','				{ $$ = sawArgDef(@1.first_line, $1, 0);  }
	| VARIABLENAME '=' scalar ','	{ $$ = sawArgDef(@1.first_line, $1, $3); }
	| VARIABLENAME '=' array  ','	{ $$ = sawArgDef(@1.first_line, $1, $3); }

defs:
	  defs def	{ $$ = appendArray($1, $2);						}
	| def		{ $$ = appendArray(nilArray(ASTTYPE_DEFS), $1);	}
	
def:
	variable_def | declaration;
	
variable_def:
	VARIABLENAME '=' expression { $$ = sawVariableDef(@1.first_line, $1, $3);	}

declaration:
	  STRING '{' string_or_var ':' proplist '}'	{ $$ = sawDeclaration(@1.first_line, $1, $3, $5); }
	| STRING '{' string_or_var ':' '}'			{ $$ = sawDeclaration(@1.first_line, $1, $3, nilArray(ASTTYPE_PROPLIST)); }

proplist:
	  proplist prop	{ $$ = appendArray($1, $2); }
	| prop			{ $$ = appendArray(nilArray(ASTTYPE_PROPLIST), $1); }
	;

prop:
	STRING ARROW value ','	{ $$ = sawProp(@1.first_line, $1, $3); }

expression:
	  value								{ $$ = $1; }
	| '(' expression ')'				{ $$ = $2; }
	| expression PLUSMINUS	expression	{ $$ = sawExpression(@1.first_line, $2, $1, $3); }
	| expression MULDIV		expression	{ $$ = sawExpression(@1.first_line, $2, $1, $3); }
	| expression COMPARISON	expression	{ $$ = sawExpression(@1.first_line, $2, $1, $3); }

value:
	  scalar		{ $$ = $1; }
	| array			{ $$ = $1; }
	| reference		{ $$ = $1; }

scalar:
	  QUOTED_STRING			{ $$ = sawQuotedString(@1.first_line, $1);			}
	| interpolated_string	{ $$ = $1;											}
	| VARIABLENAME			{ $$ = sawVariableName(@1.first_line, $1);			}
	| INT					{ $$ = sawInt(@1.first_line, $1);					}

string_or_var:
	  QUOTED_STRING			{ $$ = sawQuotedString(@1.first_line, $1);			}
	| interpolated_string	{ $$ = $1;											}
	| VARIABLENAME			{ $$ = sawVariableName(@1.first_line, $1);			}

reference:
	STRING '[' scalar ']' { $$ = sawReference(@1.first_line, $1, $3); }

array:
	  '[' arrayentries ']'	{ $$ = $2; }
	| '[' ']' 				{ $$ = nilArray(ASTTYPE_ARRAY); }

arrayentries:
	  arrayentries value ','	{ $$ = appendArray($1, $2); }
	| value ','					{ $$ = appendArray(nilArray(ASTTYPE_ARRAY), $1); }

interpolated_string:
	  INTPOL_START interpolated_string_list	{ $$ = $2; }
	| INTPOL_START							{ $$ = emptyInterpolatedString(@1.first_line); }
	  
interpolated_string_list:
	  interpolated_string_list interpolated_string_value	{ $$ = appendInterpolatedString($1, $2);		}
	| interpolated_string_value								{ $$ = appendInterpolatedString(emptyInterpolatedString(@1.first_line), $1);	}

interpolated_string_value:
	  INTPOL_VARIABLE	{ $$ = sawVariableName(@1.first_line, $1); }
	| INTPOL_TEXT 		{ $$ = sawString($1); }
	
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
