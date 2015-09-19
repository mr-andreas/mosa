%{
#include <stdio.h>

#include "lex.tab.h"  // to get the token types that we return

int line_num = 1;

%}
%%
[ \t] ;
class          { return CLASS; }
[0-9]+\.[0-9]+ { yylval.fval = atof(yytext); return FLOAT; }
[0-9]+         { yylval.ival = atoi(yytext); return INT; }
\$[a-zA-Z0-9]+   {
  yylval.sval = strdup(yytext);
  return VARIABLE;
}
[a-zA-Z0-9]+   {
  // we have to copy because we can't rely on yytext not changing underneath us:
  yylval.sval = strdup(yytext);
  return STRING;
}
'[a-zA-Z0-9]*' {
  // we have to copy because we can't rely on yytext not changing underneath us:
  yylval.sval = strdup(yytext);
  return QUOTED_STRING;
}
[\(\)\{\}:;=,]    { return yytext[0]; }
[\n]           { line_num++; }
%%
