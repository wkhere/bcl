%{
package bcl

import (
    "strconv"
)

%}

%union {
    // lexer input:
    s string

    // parsed output:
    top nTop

    blk    nBlock
    fields map[nIdent]expr
    ident  nIdent
    expr   expr
}

%token INT
%token STR
%token '{'
%token '}'
%token '('
%token ')'
%token '='
%token IDENT
%token K_VAR
%token ERR_LEX
%token EOF

%left  '+'

%%
all: vars blocks EOF        {
                                yyrcvr.lval.top = nTop{
                                    vars:   $1.top.vars,
                                    blocks: $2.top.blocks,
                                }
                                return 0
                            }

vars: /* empty */           { $$.top.vars = make(map[nIdent]expr, 2) }
    | vars K_VAR IDENT '=' expr
                            { $$.top.vars[nIdent($3.s)] = $5.expr }

blocks: /* empty */         { $$.top.blocks = nil }
    | blocks block          { $$.top.blocks = append($$.top.blocks, $2.blk) }

block:
    IDENT STR '{' fields '}' {
                                $$.blk.kind = nIdent($1.s)
                                $$.blk.name = nStrLit($2.s)
                                $$.blk.fields = $4.fields
                                // make sure nBlock has no more fields
                            }

fields: /* empty */         { $$.fields = make(map[nIdent]expr, 4) }
    | fields IDENT '=' expr { $$.fields[nIdent($2.s)] = $4.expr }

expr:
      IDENT                 { $$.expr = nVarRef(nIdent($1.s)) }
    | INT                   { $$.expr = nIntLit(atoi($1.s)) }
    | STR                   { $$.expr = nStrLit(unquote($1.s)) }
    | expr '+' expr         { $$.expr = nBinOp{"+", $1.expr, $3.expr} }
    | '(' expr ')'          { $$.expr = $2.expr }

%%

func atoi(s string) (x int) {
    x, _ = strconv.Atoi(s)
    return
}

func unquote(s string) (unquoted string) {
    unquoted, _ = strconv.Unquote(s)
    return
}
