%{
#include <stdio.h>
#define YY_DECL extern "C" int yylex()

#include "lex.tab.hpp"  // to get the token types that we return

int line_num = 1;

%}
%%
[ \t] ;
sNaZZle        { return SNAZZLE; }
type           { return TYPE; }
end            { return END; }
[0-9]+\.[0-9]+ { yylval.fval = atof(yytext); return FLOAT; }
[0-9]+         { yylval.ival = atoi(yytext); return INT; }
[a-zA-Z0-9]+   {
  // we have to copy because we can't rely on yytext not changing underneath us:
  yylval.sval = strdup(yytext);
  return STRING;
}
[\n]           { line_num++; }
.              ;
%%
