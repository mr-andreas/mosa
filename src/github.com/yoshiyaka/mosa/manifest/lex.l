%{
#include <stdio.h>

#include "parser.tab.h"  // to get the token types that we return

#define YY_USER_ACTION yylloc.first_line = yylloc.last_line = yylineno;

int line_num = 1;
int level = 0;

%}

%option yylineno
%s INBODY
%s INSTRING

%%

<INSTRING>\"						{ BEGIN(INITIAL); }
<INSTRING>\$[a-zA-Z][a-zA-Z0-9_]*	{
  printf("ITI DID HAPPEN\n");
  yylval.sval = strdup(yytext);
  return INTPOL_VARIABLE;
}
<INSTRING>[^\$\"]*						{
  yylval.sval = strdup(yytext);
  return INTPOL_TEXT;
}
\"				{ BEGIN(INSTRING); }

[ \t]  ;
\/\/.* ;
<INITIAL>class	{ return CLASS; }
<INITIAL>define	{ return DEFINE; }
<INITIAL>node	{ return NODE; }
[0-9]+\.[0-9]+	{ yylval.fval = atof(yytext); return FLOAT; }
[0-9]+			{ yylval.ival = atoi(yytext); return INT; }
=>				{ return ARROW; }
\$[a-zA-Z][a-zA-Z0-9_]* 	{
  yylval.sval = strdup(yytext);
  return VARIABLENAME;
}
[a-zA-Z][a-zA-Z0-9_]*   {
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
\{				{ ++level; BEGIN(INBODY); return '{'; }
\}				{ if(--level == 0) { BEGIN(INITIAL); } return '}'; }
[\(\):;=,[\]]	{ return yytext[0]; }
[\n]			{ line_num++; }
%%
