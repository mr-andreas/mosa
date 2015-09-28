%{
#include <stdio.h>

#include "lex.tab.h"  // to get the token types that we return

#define YY_USER_ACTION yylloc.first_line = yylloc.last_line = yylineno;

int line_num = 1;

%}

%option yylineno

%%

[ \t] ;
class			{ return CLASS; }
define			{ return DEFINE; }
node			{ return NODE; }
[0-9]+\.[0-9]+	{ yylval.fval = atof(yytext); return FLOAT; }
[0-9]+			{ yylval.ival = atoi(yytext); return INT; }
=>				{ return ARROW; }
\$[a-zA-Z0-9]+   {
  yylval.sval = strdup(yytext);
  return VARIABLENAME;
}
[a-zA-Z0-9]+   {
  // we have to copy because we can't rely on yytext not changing underneath us:
  yylval.sval = strdup(yytext);
  return STRING;
}
'[^']*' {
  // Remove the quotes at scan time
  yylval.sval = strdup(yytext+1);
  yylval.sval[strlen(yylval.sval)-1] = '\0';
  return QUOTED_STRING;
}
[\(\)\{\}:;=,[\]]    { return yytext[0]; }
[\n]           { line_num++; }
%%
