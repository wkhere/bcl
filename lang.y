%{
package bcl

import (
    "strconv"
)

type strpos struct {
    s   string
    pos pos
}

%}

%union {
    // lexer input:
    t strpos

    // parsed output:
    top  nTop
    blk  nBlock
    expr expr
}

%token tINT
%token tFLOAT
%token tSTR
%token '{'
%token '}'
%token '('
%token ')'
%token '='
%token tIDENT
%token tVAR
%token tTRUE
%token tFALSE
%token tERR
%token tEOF

%left tEQ tNE
%left '<' tLE '>' tGE
%left  '+' '-'
%left  '*' '/'
%right tNOT

%%
all: vars blocks tEOF        {
                                yyrcvr.lval.top = nTop{
                                    vars:   $1.top.vars,
                                    blocks: $2.top.blocks,
                                    pos:    0,
                                }
                                return 0
                            }

vars: /* empty */           { $$.top.vars = make(map[ident]expr, 2) }
    | vars tVAR tIDENT '=' expr
                            { $$.top.vars[ident($3.t.s)] = $5.expr }

blocks: /* empty */         { $$.top.blocks = nil }
    | blocks block          { $$.top.blocks = append($$.top.blocks, $2.blk) }

block:
    tIDENT tSTR '{' fields '}' {
                                $$.blk = nBlock{
                                    typ:    ident($1.t.s),
                                    name:   ident(unquote($2.t.s)),
                                    fields: $4.blk.fields,
                                    pos:    $1.t.pos,
                                }
                            }

fields: /* empty */         { $$.blk.fields = make(map[ident]expr, 4) }
    | fields tIDENT '=' expr { $$.blk.fields[ident($2.t.s)] = $4.expr }

expr:
      tIDENT                { $$.expr = nVarRef{ident($1.t.s), $1.t.pos} }
    | tINT                  { $$.expr = nIntLit{atoi($1.t.s),  $1.t.pos} }
    | tFLOAT                { $$.expr = nFloatLit{atof($1.t.s),  $1.t.pos} }
    | tSTR                  { $$.expr = nStrLit{unquote($1.t.s), $1.t.pos} }
    | bool_lit              { $$.expr = $1.expr }
    | expr tEQ expr         { $$.expr = nBinOp{"==", $1.expr, $3.expr} }
    | expr tNE expr         { $$.expr = nBinOp{"!=", $1.expr, $3.expr} }
    | expr '<' expr         { $$.expr = nBinOp{"<",  $1.expr, $3.expr} }
    | expr tLE expr         { $$.expr = nBinOp{"<=", $1.expr, $3.expr} }
    | expr '>' expr         { $$.expr = nBinOp{">",  $1.expr, $3.expr} }
    | expr tGE expr         { $$.expr = nBinOp{">=", $1.expr, $3.expr} }
    | expr '+' expr         { $$.expr = nBinOp{"+", $1.expr, $3.expr} }
    | expr '-' expr         { $$.expr = nBinOp{"-", $1.expr, $3.expr} }
    | expr '*' expr         { $$.expr = nBinOp{"*", $1.expr, $3.expr} }
    | expr '/' expr         { $$.expr = nBinOp{"/", $1.expr, $3.expr} }
    | '+' expr %prec tNOT   { $$.expr = $2.expr }    /* NOP */
    | '-' expr %prec tNOT   { $$.expr = nUnOp{"-",   $2.expr} }
    | tNOT expr             { $$.expr = nUnOp{"not", $2.expr} }
    | '(' expr ')'          { $$.expr = $2.expr }

bool_lit:
      tTRUE                { $$.expr = nBoolLit{true,  $1.t.pos} }
    | tFALSE               { $$.expr = nBoolLit{false, $1.t.pos} }
%%

func atoi(s string) int {
    x, _ := strconv.ParseInt(s, 0, 64)
    return int(x)
}

func atof(s string) float64 {
    x, _ := strconv.ParseFloat(s, 64)
    return x
}

func unquote(s string) (unquoted string) {
    unquoted, _ = strconv.Unquote(s)
    return
}
