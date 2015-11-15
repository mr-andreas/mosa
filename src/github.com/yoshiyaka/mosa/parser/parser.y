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
%token IF ELSE
%token <ival> BOOLTRUE BOOLFALSE
%token <sval> PLUSMINUS // + -
%token <sval> MULDIV // * /
%token <sval> COMPARISON // == > < >= <=
%token <sval> QUOTED_STRING
%token INTPOL_START
%token <sval> INTPOL_TEXT
%token <sval> INTPOL_VARIABLE

%left COMPARISON
%left PLUSMINUS
%left MULDIV

%type <gohandle> define
%type <gohandle> class
%type <gohandle> node
%type <gohandle> block
%type <gohandle> statement statements
%type <gohandle> ifstmt
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
	  NODE QUOTED_STRING block	{ $$ = sawNode(@1.first_line, $2, $3); }

class:
	  CLASS STRING optional_arg_defs block { $$ = newClass(@1.first_line, $2, $3, $4); }

block:
	  '{' statements '}' 	{ $$ = sawBlock(@1.first_line, $2); }
	| '{' '}'				{ $$ = sawBlock(@1.first_line, nilArray(ASTTYPE_STMTS)); }

statements:
	  statements statement	{ $$ = appendArray($1, $2); }
	| statement				{ $$ = appendArray(nilArray(ASTTYPE_STMTS), $1); }

statement:
	  variable_def | declaration | ifstmt;

define:
	DEFINE STRING STRING define_arg_defs block {
		$$ = sawDefine(@1.first_line, $2, $3, $4, $5);
		if($$ == -1) {
			yyerror("Expected 'single' or 'multiple' after define");
			YYABORT;
		}
	}

define_arg_defs:
	  '(' ')'			{ $$ = nilArray(ASTTYPE_ARGDEFS); }
	| '(' arg_defs ')'	{ $$ = $2; }

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
	
variable_def:
	VARIABLENAME '=' expression { $$ = sawVariableDef(@1.first_line, $1, $3);	}

declaration:
	  STRING '{' expression ':' proplist '}'	{ $$ = sawDeclaration(@1.first_line, $1, $3, $5); }
	| STRING '{' expression ':' '}'			{ $$ = sawDeclaration(@1.first_line, $1, $3, nilArray(ASTTYPE_PROPLIST)); }

ifstmt:
	  IF expression block				{ $$ = sawIf(@1.first_line, $2, $3, 0);  }
	| IF expression block ELSE block	{ $$ = sawIf(@1.first_line, $2, $3, $5); }

proplist:
	  proplist prop	{ $$ = appendArray($1, $2); }
	| prop			{ $$ = appendArray(nilArray(ASTTYPE_PROPLIST), $1); }
	;

prop:
	STRING ARROW expression ','	{ $$ = sawProp(@1.first_line, $1, $3); }

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
	  QUOTED_STRING			{ $$ = sawQuotedString(@1.first_line, $1);	}
	| interpolated_string	{ $$ = $1;									}
	| VARIABLENAME			{ $$ = sawVariableName(@1.first_line, $1);	}
	| INT					{ $$ = sawInt(@1.first_line, $1);			}
	| BOOLTRUE				{ $$ = sawBoolTrue(); 						}
	| BOOLFALSE				{ $$ = sawBoolFalse();						}

reference:
	STRING '[' scalar ']' { $$ = sawReference(@1.first_line, $1, $3); }

array:
	  '[' arrayentries ']'	{ $$ = $2; }
	| '[' ']' 				{ $$ = nilArray(ASTTYPE_ARRAY); }

arrayentries:
	  arrayentries expression ','	{ $$ = appendArray($1, $2); }
	| expression ','				{ $$ = appendArray(nilArray(ASTTYPE_ARRAY), $1); }

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
