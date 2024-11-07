#!/usr/bin/env python3

import subprocess
from os import system
from sys import stderr, exit

tests = [
    [ '0',   '', ''],
    [ '0.1', 'eval "expr that is discarded"', ''],
    [ '0.2', 'eval nil', ''],
    [ '0.3', 'eval "anything"', ''],
    [ '1',   'var a; print not a', 'true'],
    [ '2',   'var a; print a==nil', 'true'],
    [ '3',   'var a=1; eval a=nil; print a==nil', 'true'],
    [ '4',   'print 1+1',    '2'],
    [ '5',   'print 1+2.14', '3.14'],
    [ '6',   'print 123/2-50+2*8',     '27'],
    [ '6.1', 'print 123.0/2-50+2*8',   '27.5'],
    [ '6.2', 'print 123/2.0-50+2*8',   '27.5'],
    [ '6.3', 'print 123.0/2.0-50+2*8', '27.5'],
    [ '6.4', 'print 123/2-50+2.0*8',   '27'],
    [ '6.5', 'print 123/2-50+2*8.0',   '27'],
    [ '6.6', 'print 123/2-50+2.0*8.0', '27'],
    [ '6.7', 'print 123/2-50.0+2*8',   '27'],
    [ '7',   'print 1-2',  '-1'],
    [ '8',   'print 1--2',  '3'],
    [ '9',   'print 1- +1', '0'],
    ['10',   'print 1+ +1',     '2'],
    ['10.1', 'print 1.0+ +1',   '2'],
    ['10.2', 'print 1+ +1.0',   '2'],
    ['10.3', 'print 1.0+ +1.0', '2'],
    ['11',   'print ---10',   '-10'],
    ['11.1', 'print ---10.0', '-10'],
    ['12',   'print 1==1',     'true'],
    ['12.1', 'print 1.0==1',   'true'],
    ['12.2', 'print 1==1.0',   'true'],
    ['12.3', 'print 1.0==1.0', 'true'],
    ['13',   'print not 1>3',     'true'],
    ['13.1', 'print not 1.0>3',   'true'],
    ['13.2', 'print not 1>3.0',   'true'],
    ['13.3', 'print not 1.0>3.0', 'true'],
    ['13.4', 'print not 3<=1',    'true'],
    ['14',   'print 1<2 and 0 or "whatever"',       'whatever'],
    ['14.1', 'print 1.0<2 and 0 or "whatever"',     'whatever'],
    ['14.2', 'print 1<2.0 and 0 or "whatever"',     'whatever'],
    ['14.3', 'print 1.0<2.0 and 0 or "whatever"',   'whatever'],
    ['14.4', 'print 1<2 and 0.0 or "whatever"',     'whatever'],
    ['14.5', 'print 1.0<2 and 0.0 or "whatever"',   'whatever'],
    ['14.6', 'print 1<2.0 and 0.0 or "whatever"',   'whatever'],
    ['14.7', 'print 1.0<2.0 and 0.0 or "whatever"', 'whatever'],
    ['15',   'print 1>2 or true and 42',    '42'],
    ['16',   'print 1<10/5 and 127*-1+154', '27'],
    ['17',   'print "q"*2', 'qq'],
    ['18',   'print "q"+"x"', 'qx'],
    ['19',   'print "q"+3', 'q3'],
    ['20',   'print "q"+3.14', 'q3.14'],
    ['21',   'print "q"=="q"', 'true'],
    ['22',   'print "q"!="p"', 'true'],
    ['23',   'print "p"<"q"',  'true'],
    ['24',   'print "q">"p"',  'true'],
    ['25',   'print not (false or true)',  'false'],
    ['25.1', 'print not (false and true)', 'true'],
    ['25.2', 'print "" or 42',       '42'],
    ['26',   'var a=100; var b=a-90; print -b', '-10'],
    ['27',   'var a=1; var b=2; print a+b',  '3'],
    ['28',   'var a=1; eval a=a+1; print a', '2'],
    ['29',   'var a=1; print a=a+1',         '2'],
    ['30',   'def blk {print TYPE}',         'blk'],
    ['31',   'def blk "foo" {print TYPE+"."+NAME}', 'blk.foo'],
    ['32',   'var a=1; def blk {print TYPE}',   'blk'],
    ['33',   'print 1; def blk {print TYPE}',   '1\nblk'],
    ['34',   'var a=1; def blk {print TYPE+a}', 'blk1'],
    ['35',   'def blk {var a=1; print TYPE+a}', 'blk1'],
    ['36',   'var x=1; def blk {var a=1+x; print TYPE+a}',   'blk2'],
    ['37',   'def blk {var a=1; var b=2; print TYPE+(a+b)}', 'blk3'],
    ['38',   'var x=5; def blk {var a=1; var b=2; print TYPE+(a+b+x)}', 'blk8'],
    ['39',   'def b1{var x=1; def b2 {var a=2; print TYPE+(a+x)}}',     'b23'],
    ['40',   'var a=5; def blk {var a=a+1; print TYPE+a}',   'blk6'],
    ['41',   'def b1{var a=5; def b2 {var a=a+1; print TYPE+a} print TYPE+a}', 'b26\nb15'],
    ['42',   'def b1{var a=5; def b2 {eval a=a+1; print TYPE+a} print TYPE+a}','b26\nb16'],
    ['43',   'def b1{ print TYPE; def b2{print TYPE} }; def b3{print TYPE}', 'b1\nb2\nb3'],
    ['44',   'var a; var b; eval a=1+(b=2); print a', '3'],
    ['45',   'var a; var b; print a=1+(b=2)',         '3'],
    ['46',   'def x {42}', ''],
    ['47',   'def x {var x=42; x+1; print x}', '42'],
    ['48',   'def x {a=1+(b=2); print a}', '3'],
    ['49',   'def x {print a=1+(b=2)}',    '3'],
    ['50',   'def x {print (a=1)+(b=2)}',  '3'],

    ['51',   '', '== /dev/stdin ==\n0000    1:1  RET', 'disasm'],
    ['52',   'eval nil',
        '== /dev/stdin ==\n'
        '0000    1:9  NIL\n'
        '0001      |  POP\n'
        '0002      |  RET',
        'disasm'
    ],
    ['53',   'eval 42',
        "== /dev/stdin ==\n"
        "0000    1:8  CONST         0 '42'\n"
        "0002      |  POP\n"
        "0003      |  RET",
        'disasm'
    ],
    ['54',   'def b {}',
        "== /dev/stdin ==\n"
        "0000    1:8  DEFBLOCK      0 'b'\t   1 ''\n"
        "0003    1:9  ENDBLOCK\n"
        "0004      |  RET",
        'disasm'
    ],

    ['55',   '1', '',           "err: expected statement"],
    ['55.1', '=1', '',          "err: expected statement"],
    ['56',   'print', '',       "err: at end: expected expression"],
    ['56.1', 'print print', '', "err: at 'print': expected expression"],
    ['56.2', 'print =', '',     "err: at '=': expected expression"],
    ['57',   'eval', '',        "err: expected expression"],
    ['58',   'eval (1', '',     "err: expected ')'"],
    ['59',   'def 1', '',         "err: at '1': expected block type"],
    ['60',   'def b {', '',       "err: at end: expected '}'"],
    ['61.1', 'def b x {}', '',    "err: at 'x': expected '{'"],
    ['61.2', 'def b 0 {}', '',    "err: at '0': expected '{'"],
    ['62.1', 'def b "x" {', '',   "err: at end: expected '}'"],
    ['62.2', 'def b "x" { z', '', "err: at end: expected '}'"],
    ['63.1', 'eval 1 =', '',      "err: at '=': invalid assignment target"],
    ['63.2', 'eval "a" = ', '',   "err: at '=': invalid assignment target"],
    ['63.3', 'eval false = ', '', "err: at '=': invalid assignment target"],
    ['64',   'eval a=42', '',     "err: at '=': invalid assignment target"],
    ['65',   'var a; var a', '',  "err: at 'a': variable with this name already present"],
    ['66',   'eval a', '',        "err: at 'a': undefined variable"],

    ['67.1',  'print -1',            "-1"],
    ['67.2',  'print -1.2',          "-1.2"],
    ['67.3',  'print -(1)',          "-1"],
    ['67.4',  'print -(1.2)',        "-1.2"],
    ['67.5',  'var a=1;   print -a', "-1"],
    ['67.6',  'var a=1.2; print -a', "-1.2"],
    ['67.7',  'print -true', '',     "err: NEG: invalid type: bool, expected number"],
    ['67.8',  'print -"abc"', '',    "err: NEG: invalid type: string, expected number"],
    ['67.9',  'print -nil', '',      "err: NEG: invalid type: nil, expected number"],

    ['68.1',  'print +1',            "1"],
    ['68.2',  'print +1.2',          "1.2"],
    ['68.3',  'print +(1)',          "1"],
    ['68.4',  'print +(1.2)',        "1.2"],
    ['67.5',  'var a=1;   print +a', "1"],
    ['67.6',  'var a=1.2; print +a', "1.2"],
    ['68.7',  'print +true', '',     "err: UNPLUS: invalid type: bool, expected number"],
    ['68.8',  'print +"abcd"', '',   "err: UNPLUS: invalid type: string, expected number"],
    ['68.9',  'print +nil', '',      "err: UNPLUS: invalid type: nil, expected number"],

    ['69',    'print *1', '',        "err: at '*': expected expression"],
    ['70',    'print /1', '',        "err: at '/': expected expression"],

    ['71.1',  'print 1+2',           '3'],
    ['71.2',  'print 1+2.5',         '3.5'],
    ['71.3',  'print 1+true', '',    "err: ADD: invalid types: int, bool"],
    ['71.4',  'print 1+"ab"', '',    "err: ADD: invalid types: int, string"],
    ['71.5',  'print 1+nil', '',     "err: ADD: invalid types: int, nil"],
    ['72.1',  'print 1.2+5',         '6.2'],
    ['72.2',  'print 1.2+3.5',       '4.7'],
    ['72.3',  'print 1.2+true', '',  "err: ADD: invalid types: float, bool"],
    ['72.4',  'print 1.2+"ab"', '',  "err: ADD: invalid types: float, string"],
    ['72.5',  'print 1.2+nil', '',   "err: ADD: invalid types: float, nil"],
    ['73.1',  'print true+1', '',    "err: ADD: invalid types: bool, int"],
    ['73.2',  'print true+1.2', '',  "err: ADD: invalid types: bool, float"],
    ['73.3',  'print true+true', '', "err: ADD: invalid types: bool, bool"],
    ['73.4',  'print true+"ab"', '', "err: ADD: invalid types: bool, string"],
    ['73.5',  'print true+nil', '',  "err: ADD: invalid types: bool, nil"],
    ['74.1',  'print "ab"+1',        'ab1'],
    ['74.2',  'print "ab"+1.2',      'ab1.2'],
    ['74.3',  'print "ab"+true', '', "err: ADD: invalid types: string, bool"],
    ['74.4',  'print "ab"+"cd"',     'abcd'],
    ['74.5',  'print "ab"+nil',      'ab'],
    ['75.1',  'print nil+1', '',     "err: ADD: invalid types: nil, int"],
    ['75.2',  'print nil+1.2', '',   "err: ADD: invalid types: nil, float"],
    ['75.3',  'print nil+true', '',  "err: ADD: invalid types: nil, bool"],
    ['75.4',  'print nil+"ab"', '',  "err: ADD: invalid types: nil, string"],
    ['75.5',  'print nil+nil', '',   "err: ADD: invalid types: nil, nil"],

    ['76.1',  'print 1-2',           '-1'],
    ['76.2',  'print 1-2.5',         '-1.5'],
    ['76.3',  'print 1-true', '',    "err: SUB: invalid types: int, bool"],
    ['76.4',  'print 1-"ab"', '',    "err: SUB: invalid types: int, string"],
    ['76.5',  'print 1-nil', '',     "err: SUB: invalid types: int, nil"],
    ['77.1',  'print 1.5-2',         '-0.5'],
    ['77.2',  'print 1.0-2.5',       '-1.5'],
    ['77.3',  'print 1.2-true', '',  "err: SUB: invalid types: float, bool"],
    ['77.4',  'print 1.2-"ab"', '',  "err: SUB: invalid types: float, string"],
    ['77.5',  'print 1.2-nil', '',   "err: SUB: invalid types: float, nil"],
    ['78.1',  'print true-1', '',    "err: SUB: invalid types: bool, int"],
    ['78.2',  'print true-1.2', '',  "err: SUB: invalid types: bool, float"],
    ['78.3',  'print true-true', '', "err: SUB: invalid types: bool, bool"],
    ['78.4',  'print true-"ab"', '', "err: SUB: invalid types: bool, string"],
    ['78.5',  'print true-nil', '',  "err: SUB: invalid types: bool, nil"],
    ['79.1',  'print "ab"-1', '',    "err: SUB: invalid types: string, int"],
    ['79.2',  'print "ab"-1.2', '',  "err: SUB: invalid types: string, float"],
    ['79.3',  'print "ab"-true', '', "err: SUB: invalid types: string, bool"],
    ['79.4',  'print "ab"-"cd"', '', "err: SUB: invalid types: string, string"],
    ['79.5',  'print "ab"-nil', '',  "err: SUB: invalid types: string, nil"],
    ['80.1',  'print nil-1', '',     "err: SUB: invalid types: nil, int"],
    ['80.2',  'print nil-1.2', '',   "err: SUB: invalid types: nil, float"],
    ['80.3',  'print nil-true', '',  "err: SUB: invalid types: nil, bool"],
    ['80.4',  'print nil-"ab"', '',  "err: SUB: invalid types: nil, string"],
    ['80.5',  'print nil-nil', '',   "err: SUB: invalid types: nil, nil"],

    ['81.1',  'print 1*2',           '2'],
    ['81.2',  'print 1*2.5',         '2.5'],
    ['81.3',  'print 1*true', '',    "err: MUL: invalid types: int, bool"],
    ['81.4',  'print 1*"ab"', '',    "err: MUL: invalid types: int, string"],
    ['81.5',  'print 1*nil', '',     "err: MUL: invalid types: int, nil"],
    ['82.1',  'print 1.4*2',         '2.8'],
    ['82.2',  'print 1.0*2.5',       '2.5'],
    ['82.3',  'print 1.2*true', '',  "err: MUL: invalid types: float, bool"],
    ['82.4',  'print 1.2*"ab"', '',  "err: MUL: invalid types: float, string"],
    ['82.5',  'print 1.2*nil', '',   "err: MUL: invalid types: float, nil"],
    ['83.1',  'print true*1', '',    "err: MUL: invalid types: bool, int"],
    ['83.2',  'print true*1.2', '',  "err: MUL: invalid types: bool, float"],
    ['83.3',  'print true*true', '', "err: MUL: invalid types: bool, bool"],
    ['83.4',  'print true*"ab"', '', "err: MUL: invalid types: bool, string"],
    ['83.5',  'print true*nil', '',  "err: MUL: invalid types: bool, nil"],
    ['84.1',  'print "ab"*2',        'abab'],
    ['84.2',  'print "ab"*1.2', '',  "err: MUL: invalid types: string, float"],
    ['84.3',  'print "ab"*true', '', "err: MUL: invalid types: string, bool"],
    ['84.4',  'print "ab"*"cd"', '', "err: MUL: invalid types: string, string"],
    ['84.5',  'print "ab"*nil', '',  "err: MUL: invalid types: string, nil"],
    ['85.1',  'print nil*1', '',     "err: MUL: invalid types: nil, int"],
    ['85.2',  'print nil*1.2', '',   "err: MUL: invalid types: nil, float"],
    ['85.3',  'print nil*true', '',  "err: MUL: invalid types: nil, bool"],
    ['85.4',  'print nil*"ab"', '',  "err: MUL: invalid types: nil, string"],
    ['85.5',  'print nil*nil', '',   "err: MUL: invalid types: nil, nil"],

    ['86.1',  'print 1/2',           '0'],
    ['86.2',  'print 1/2.0',         '0.5'],
    ['86.3',  'print 1/true', '',    "err: DIV: invalid types: int, bool"],
    ['86.4',  'print 1/"ab"', '',    "err: DIV: invalid types: int, string"],
    ['86.5',  'print 1/nil', '',     "err: DIV: invalid types: int, nil"],
    ['87.1',  'print 1.0/2',         '0.5'],
    ['87.2',  'print 1.0/2.0',       '0.5'],
    ['87.3',  'print 1.2/true', '',  "err: DIV: invalid types: float, bool"],
    ['87.4',  'print 1.2/"ab"', '',  "err: DIV: invalid types: float, string"],
    ['87.5',  'print 1.2/nil', '',   "err: DIV: invalid types: float, nil"],
    ['88.1',  'print true/1', '',    "err: DIV: invalid types: bool, int"],
    ['88.2',  'print true/1.2', '',  "err: DIV: invalid types: bool, float"],
    ['88.3',  'print true/true', '', "err: DIV: invalid types: bool, bool"],
    ['88.4',  'print true/"ab"', '', "err: DIV: invalid types: bool, string"],
    ['88.5',  'print true/nil', '',  "err: DIV: invalid types: bool, nil"],
    ['89.1',  'print "ab"/2', '',    "err: DIV: invalid types: string, int"],
    ['89.2',  'print "ab"/1.2', '',  "err: DIV: invalid types: string, float"],
    ['89.3',  'print "ab"/true', '', "err: DIV: invalid types: string, bool"],
    ['89.4',  'print "ab"/"cd"', '', "err: DIV: invalid types: string, string"],
    ['89.5',  'print "ab"/nil', '',  "err: DIV: invalid types: string, nil"],
    ['90.1',  'print nil/1', '',     "err: DIV: invalid types: nil, int"],
    ['90.2',  'print nil/1.2', '',   "err: DIV: invalid types: nil, float"],
    ['90.3',  'print nil/true', '',  "err: DIV: invalid types: nil, bool"],
    ['90.4',  'print nil/"ab"', '',  "err: DIV: invalid types: nil, string"],
    ['90.5',  'print nil/nil', '',   "err: DIV: invalid types: nil, nil"],

    ['91.1',  'print 1==1',           'true'],
    ['91.2',  'print 1==1.0',         'true'],
    ['91.3',  'print 1==true',        'false'],
    ['91.4',  'print 1=="ab"',        'false'],
    ['91.5',  'print 1==nil',         'false'],
    ['92.1',  'print 1.0==1',         'true'],
    ['92.2',  'print 1.0==1.0',       'true'],
    ['92.3',  'print 1.2==true',      'false'],
    ['92.4',  'print 1.2=="ab"',      'false'],
    ['92.5',  'print 1.2==nil',       'false'],
    ['93.1',  'print true==1',        'false'],
    ['93.2',  'print true==1.2',      'false'],
    ['93.3',  'print true==true',     'true'],
    ['93.4',  'print true=="ab"',     'false'],
    ['93.5',  'print true==nil',      'false'],
    ['94.1',  'print "ab"==2',        'false'],
    ['94.2',  'print "ab"==1.2',      'false'],
    ['94.3',  'print "ab"==true',     'false'],
    ['94.4',  'print "ab"=="cd"',     'false'],
    ['94.5',  'print "ab"==nil',      'false'],
    ['95.1',  'print nil==1',         'false'],
    ['95.2',  'print nil==1.2',       'false'],
    ['95.3',  'print nil==true',      'false'],
    ['95.4',  'print nil=="ab"',      'false'],
    ['95.5',  'print nil==nil',       'true'],

    ['96.1',  'print 1<2',           'true'],
    ['96.2',  'print 1<2.0',         'true'],
    ['96.3',  'print 1<true', '',    "err: LT: invalid types: int, bool"],
    ['96.4',  'print 1<"ab"', '',    "err: LT: invalid types: int, string"],
    ['96.5',  'print 1<nil', '',     "err: LT: invalid types: int, nil"],
    ['97.1',  'print 1.0<2',         'true'],
    ['97.2',  'print 1.0<2.0',       'true'],
    ['97.3',  'print 1.2<true', '',  "err: LT: invalid types: float, bool"],
    ['97.4',  'print 1.2<"ab"', '',  "err: LT: invalid types: float, string"],
    ['97.5',  'print 1.2<nil', '',   "err: LT: invalid types: float, nil"],
    ['98.1',  'print true<1', '',    "err: LT: invalid types: bool, int"],
    ['98.2',  'print true<1.0', '',  "err: LT: invalid types: bool, float"],
    ['98.3',  'print true<true', '', "err: LT: invalid types: bool, bool"],
    ['98.4',  'print true<"ab"', '', "err: LT: invalid types: bool, string"],
    ['98.5',  'print true<nil', '',  "err: LT: invalid types: bool, nil"],
    ['99.1',  'print "ab"<2', '',    "err: LT: invalid types: string, int"],
    ['99.2',  'print "ab"<1.2', '',  "err: LT: invalid types: string, float"],
    ['99.3',  'print "ab"<true', '', "err: LT: invalid types: string, bool"],
    ['99.4',  'print "ab"<"cd"',     'true'],
    ['99.5',  'print "ab"<nil', '',  "err: LT: invalid types: string, nil"],
    ['100.1', 'print nil<1', '',     "err: LT: invalid types: nil, int"],
    ['100.2', 'print nil<1.2', '',   "err: LT: invalid types: nil, float"],
    ['100.3', 'print nil<true', '',  "err: LT: invalid types: nil, bool"],
    ['100.4', 'print nil<"ab"', '',  "err: LT: invalid types: nil, string "],
    ['100.5', 'print nil<nil', '',   "err: LT: invalid types: nil, nil"],

    ['101.1',  'print 1>2',           'false'],
    ['101.2',  'print 1>2.0',         'false'],
    ['101.3',  'print 1>true', '',    "err: GT: invalid types: int, bool"],
    ['101.4',  'print 1>"ab"', '',    "err: GT: invalid types: int, string"],
    ['101.5',  'print 1>nil', '',     "err: GT: invalid types: int, nil"],
    ['102.1',  'print 1.0>2',         'false'],
    ['102.2',  'print 1.0>2.0',       'false'],
    ['102.3',  'print 1.2>true', '',  "err: GT: invalid types: float, bool"],
    ['102.4',  'print 1.2>"ab"', '',  "err: GT: invalid types: float, string"],
    ['102.5',  'print 1.2>nil', '',   "err: GT: invalid types: float, nil"],
    ['103.1',  'print true>1', '',    "err: GT: invalid types: bool, int"],
    ['103.2',  'print true>1.0', '',  "err: GT: invalid types: bool, float"],
    ['103.3',  'print true>true', '', "err: GT: invalid types: bool, bool"],
    ['103.4',  'print true>"ab"', '', "err: GT: invalid types: bool, string"],
    ['103.5',  'print true>nil', '',  "err: GT: invalid types: bool, nil"],
    ['104.1',  'print "ab">2', '',    "err: GT: invalid types: string, int"],
    ['104.2',  'print "ab">1.2', '',  "err: GT: invalid types: string, float"],
    ['104.3',  'print "ab">true', '', "err: GT: invalid types: string, bool"],
    ['104.4',  'print "ab">"cd"',     'false'],
    ['104.5',  'print "ab">nil', '',  "err: GT: invalid types: string, nil"],
    ['105.1',  'print nil>1', '',     "err: GT: invalid types: nil, int"],
    ['105.2',  'print nil>1.2', '',   "err: GT: invalid types: nil, float"],
    ['105.3',  'print nil>true', '',  "err: GT: invalid types: nil, bool"],
    ['105.4',  'print nil>"ab"', '',  "err: GT: invalid types: nil, string "],
    ['105.5',  'print nil>nil', '',   "err: GT: invalid types: nil, nil"],

    ['106.1',  'print not 0',        'true'],
    ['106.2',  'print not 0.0',      'true'],
    ['106.3',  'print not false',    'true'],
    ['106.4',  'print not ""',       'true'],
    ['106.5',  'print not nil',      'true'],

    ['107.1',  'print 1 or 2',         '1'],
    ['107.2',  'print 1 or 2.0',       '1'],
    ['107.3',  'print 1 or true',      '1'],
    ['107.4',  'print 1 or "ab"',      '1'],
    ['107.5',  'print 1 or nil',       '1'],
    ['108.1',  'print 1.2 or 2',       '1.2'],
    ['108.2',  'print 1.2 or 2.0',     '1.2'],
    ['108.3',  'print 1.2 or true',    '1.2'],
    ['108.4',  'print 1.2 or "ab"',    '1.2'],
    ['108.5',  'print 1.2 or nil',     '1.2' ],
    ['109.1',  'print true or 1',      'true'],
    ['109.2',  'print true or 1.0',    'true'],
    ['109.3',  'print true or true',   'true'],
    ['109.4',  'print true or "ab"',   'true'],
    ['109.5',  'print true or nil',    'true'],
    ['110.1',  'print "ab" or 2',      'ab'],
    ['110.2',  'print "ab" or 1.2',    'ab'],
    ['110.3',  'print "ab" or true',   'ab'],
    ['110.4',  'print "ab" or "cd"',   'ab'],
    ['110.5',  'print "ab" or nil',    'ab'],
    ['112.1',  'print nil or 1',       '1'],
    ['112.2',  'print nil or 1.2',     '1.2'],
    ['112.3',  'print nil or true',    'true'],
    ['112.4',  'print nil or "ab"',    'ab'],
    ['112.5',  'print nil or nil',     '<nil>'],

    ['113.1',  'print 1 and 2',         '2'],
    ['113.2',  'print 1 and 2.5',       '2.5'],
    ['113.3',  'print 1 and true',      'true'],
    ['113.4',  'print 1 and "ab"',      'ab'],
    ['113.5',  'print 1 and nil',       '<nil>'],
    ['114.1',  'print 1.2 and 2',       '2'],
    ['114.2',  'print 1.2 and 2.5',     '2.5'],
    ['114.3',  'print 1.2 and true',    'true'],
    ['114.4',  'print 1.2 and "ab"',    'ab'],
    ['114.5',  'print 1.2 and nil',     '<nil>'],
    ['115.1',  'print true and 2',      '2'],
    ['115.2',  'print true and 2.5',    '2.5'],
    ['115.3',  'print true and true',   'true'],
    ['115.4',  'print true and "ab"',   'ab'],
    ['115.5',  'print true and nil',    '<nil>'],
    ['116.1',  'print "ab" and 2',      '2'],
    ['116.2',  'print "ab" and 2.5',    '2.5'],
    ['116.3',  'print "ab" and true',   'true'],
    ['116.4',  'print "ab" and "cd"',   'cd'],
    ['116.5',  'print "ab" and nil',    '<nil>'],
    ['117.1',  'print nil and 2',       '<nil>'],
    ['117.2',  'print nil and 2.5',     '<nil>'],
    ['117.3',  'print nil and true',    '<nil>'],
    ['117.4',  'print nil and "ab"',    '<nil>'],
    ['117.5',  'print nil and nil',     '<nil>'],

    ['120',    'def b{x}', '',  "err: 'x' not resolved as var or field"],
    ['121.1',  'print 1/0', '', "err: division by int zero"],
    ['121.2',  'print 1/0.0',   '+Inf'],

    ['122.1',  f'print  {(1<<31)-1}-1',  f'{ (1<<31)-2}'],
    ['122.2',  f'print -{(1<<31)-1}+1',  f'{-(1<<31)+2}'],

    ['123.1',  'var a; print "foo"+(a=1); print a', 'foo1\n1'],
    ['123.2',  'var a; print 2+(a=1); print a',     '3\n1'],
    ['123.3',  'def b{print "foo"+(a=1); print a}', 'foo1\n1'],
    ['123.4',  'def b{var a; print "foo"+(a=1); print a}', 'foo1\n1'],
]

tests_64b = [
    ['122.1-64', f'print  {(1<<63)-1}-1',  f'{ (1<<63)-2}'],
    ['122.2-64', f'print -{(1<<63)-1}+1',  f'{-(1<<63)+2}'],
]


def err_match(opt):
     m = next((s for s in opt if s == 'err' or s.startswith('err:')), None)
     if m:
        if m == 'err': return m
        return m[4:].strip() or 'err'


def perr(*args):
    print(*args, file=stderr)


def run_tests():
    cmd = './bcl'.split()
    bindesc = subprocess.getoutput('file ./bcl')
    tests_extra = tests_64b if 'ELF 64-bit' in bindesc else [] 

    fail = False
    for i, prog, exp, *opt in tests + tests_extra:
        cmd2 = cmd.copy()
        if 'disasm' in opt: cmd2.append('--disasm')

        proc = subprocess.run(cmd2, input=prog, text=True, capture_output=True)

        if proc.returncode == 0:
            if m := err_match(opt):
                perr(f'#{i} no error when expected one matching {m!r}')
                fail = True

            elif (res := proc.stdout.rstrip()) != exp:
                fail = True
                perr(f'#{i} mismatch: have {res!r}, want {exp!r}')

        else:
            msg = proc.stderr.rstrip()
            msg = removesuffix(msg, "\ncombined errors from parse")
            if m := err_match(opt):
                if m == 'err': continue
                if m not in msg:
                    perr(f'#{i} error mismatch: have {msg!r}, want matching {m!r}')
                    fail = True
            else:
                perr(f'#{i} unexpected error: {msg!r}')
                fail = True

    if fail: exit(1)


def removesuffix(s, suffix):
    if s.endswith(suffix):
        return s[:-len(suffix)]
    else:
        return s


def q(s): return s.encode('unicode-escape').decode()


part1 = r"""// Code generated by "./test.py generate"; DO NOT EDIT.

package bcl_test

import (
    "bytes"
    "strings"
    "testing"

    "github.com/wkhere/bcl"
)

type stdinBuf struct{ *strings.Reader }

func (stdinBuf) Name() string { return "/dev/stdin" }
func (stdinBuf) Close() error { return nil }

func TestInterpretFromPy(t *testing.T) {
    var tab = []struct {
        name, input, output string
        disasm              bool
        errWanted           bool
        errMatch            string
    }{"""

part2 = r""" }

    for _, tc := range tab {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            inp := stdinBuf{strings.NewReader(tc.input)}        
            out := new(bytes.Buffer)
            log := new(bytes.Buffer)

            _, err := bcl.InterpretFile(
                inp,
                bcl.OptDisasm(tc.disasm),
                bcl.OptOutput(out), bcl.OptLogger(log),
            )

            switch {
            case err != nil && !tc.errWanted:
                t.Errorf("unexpected error: %s", relevantError(err, log))

            case err != nil && tc.errWanted:
                rerr := relevantError(err, log)
                if !strings.Contains(rerr, tc.errMatch) {
                    t.Errorf("error mismatch\nhave: %s\nwant matching: %s",
                        rerr, tc.errMatch,
                    )
                }

            case err == nil && tc.errWanted && tc.errMatch == "":
                t.Errorf("no error when expecting one")

            case err == nil && tc.errWanted && tc.errMatch != "":
                t.Errorf("no error when expecting one matching: %s", tc.errMatch)

            case err == nil && !tc.errWanted:
                s := strings.TrimRight(out.String(), "\n")
                if s != tc.output {
                    t.Errorf("mismatch:\nhave: %s\nwant: %s", s, tc.output)
                }
            }
        })
    }
}

func relevantError(err error, buf *bytes.Buffer) string {
    if s := err.Error(); strings.HasPrefix(s, "combined errors") {
        return strings.TrimRight(buf.String(), "\n")
    } else {
        return s
    }
}
"""


def generate(target):
    arch = subprocess.getoutput('go env GOARCH')
    tests_extra = tests_64b if '64' in arch or arch in ('s390x', 'wasm') else []

    with open(target, 'w') as f:
        print(part1, file=f)

        for (i, inp, outp, *opt) in tests + tests_extra:
            print(f'\t\t{{`{i}`, `{inp}`, "{q(outp)}", ', file=f, end='')
            print('true, ' if 'disasm' in opt else 'false, ', file=f, end='')
            if m := err_match(opt):
                print('true, ', file=f, end='')
                msg = '""' if m == 'err' else f'"{q(m)}"'
                print(msg, file=f, end='')
            else:
                print('false, ""', file=f, end='')
            print('},', file=f)

        print(part2, file=f)


if __name__ == '__main__':
    from sys import argv as args
    if len(args)>1 and args[1] == 'generate':
        if len(args)>2 and args[2].endswith('.go'):
            generate(args[2])
        else:
            raise SystemExit('generate, but what?')
    else:
        run_tests()
