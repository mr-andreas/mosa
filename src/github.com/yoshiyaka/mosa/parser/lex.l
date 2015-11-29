%{
#include <stdio.h>

#include "parser.tab.h"  // to get the token types that we return

#define YY_USER_ACTION yylloc.first_line = yylloc.last_line = yylineno;

int line_num = 1;
int level = 0;

%}

%option yylineno
/* %option debug */
%s INBODY
%s INSTRING

%%

<INSTRING>\"							{ if(level > 0) BEGIN(INBODY); else BEGIN(INITIAL); }
<INSTRING>\$[a-zA-Z][a-zA-Z0-9_]*		{
  yylval.sval = strdup(yytext);
  return INTPOL_VARIABLE;
}
<INSTRING>\$\{[a-zA-Z][a-zA-Z0-9_]*\}	{
  // Normalize ${foo} to $foo directly at lex time.
  yylval.sval = strdup(yytext+1);
  yylval.sval[0] = '$';
  yylval.sval[strlen(yylval.sval)-1] = '\0';
  return INTPOL_VARIABLE;
}
<INSTRING>[^\$"]*						{
  yylval.sval = strdup(yytext);
  return INTPOL_TEXT;
}
<INSTRING>\$							{
  yylval.sval = strdup(yytext);
  return INTPOL_TEXT;
}
\"				{ BEGIN(INSTRING); return INTPOL_START; }

[ \t]  ;
\/\/.* ;
<INITIAL>class	{ return CLASS; }
<INITIAL>define	{ return DEFINE; }
<INITIAL>node	{ return NODE; }
<INITIAL>func	{ return FUNC; }
if				{ return IF; }
else			{ return ELSE; }
true			{ return BOOLTRUE; }
false			{ return BOOLFALSE; }

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
<INITIAL,INBODY>'[^']*' {
  // Use INITIAL,INBODY here so that we don't match a quoted string inside of an
  // interpolated string.

  // Remove the quotes at scan time
  yylval.sval = strdup(yytext+1);
  yylval.sval[strlen(yylval.sval)-1] = '\0';
  return QUOTED_STRING;
}
\{				{ ++level; BEGIN(INBODY); return '{'; }
\}				{ if(--level == 0) { BEGIN(INITIAL); } return '}'; }
[\n]			{ line_num++; }
[+-]			{
  yylval.sval = strdup(yytext);
  return PLUSMINUS;
}
[*/]			{
  yylval.sval = strdup(yytext);
  return MULDIV;
}
[><]=?			{
  yylval.sval = strdup(yytext);
  return COMPARISON;
}
==				{
  yylval.sval = strdup(yytext);
  return COMPARISON;
}
!=				{
  yylval.sval = strdup(yytext);
  return COMPARISON;
}
[\(\):;=,[\]]	{ return yytext[0]; }
%%
