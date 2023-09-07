%{
package main

import (
    "strconv"
)

%}

%union {
    // lexer input:
    s string

    // parsed output:
    top nTop

    tunnel  nTunnel
    fields  map[nIdent]expr
    ident   nIdent
    expr    expr
}

%token INT
%token STR
%token '{'
%token '}'
%token '='
%token IDENT
%token K_VAR
%token K_TUNNEL
%token ERR_LEX
%token EOF

%left  '+'

%%
all: vars tunnels EOF       {
                                yyrcvr.lval.top = nTop{
                                    vars:    $1.top.vars,
                                    tunnels: $2.top.tunnels,
                                }
                                return 0
                            }

vars: /* empty */           { $$.top.vars = make(map[nIdent]expr, 2) }
    | vars K_VAR IDENT '=' expr
                            { $$.top.vars[nIdent($3.s)] = $5.expr }

tunnels: /* empty */        { $$.top.tunnels = nil }
    | tunnels tunnel        { $$.top.tunnels =
                                append($$.top.tunnels, $2.tunnel)
                            }

tunnel:
    K_TUNNEL STR '{' fields '}' {
                                $$.tunnel.name = nStrLit($2.s)
                                $$.tunnel.fields = $4.fields
                            }

fields: /* empty */         { $$.fields = make(map[nIdent]expr, 4) }
    | fields IDENT '=' expr { $$.fields[nIdent($2.s)] = $4.expr }

expr:
      IDENT                 { $$.expr = nVarRef(nIdent($1.s)) }
    | INT                   { $$.expr = nIntLit(atoi($1.s)) }
    | STR                   { $$.expr = nStrLit(unquote($1.s)) }
    | expr '+' expr         { $$.expr = nBinOp{'+', $1.expr, $3.expr} }

%%

func atoi(s string) (x int) {
    x, _ = strconv.Atoi(s)
    return
}

func unquote(s string) (unquoted string) {
    unquoted, _ = strconv.Unquote(s)
    return
}
