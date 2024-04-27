#!/usr/bin/env python3

import subprocess
from sys import stderr, exit

tests = [
    [ 0,   '', ''],
    [ 0.1, 'eval "expr that is discarded"', ''],
    [ 0.2, 'eval nil', ''],
    [ 0.3, 'eval "anything"', ''],
    [ 1, 'var a; print not a', 'true'],
    [ 2, 'var a; print a==nil', 'true'],
    [ 3, 'var a=1; eval a=nil; print a==nil', 'true'],
    [ 4, 'print 1+1', '2'],
    [ 5, 'print 1+2.14', '3.14'],
    [ 6,   'print 123/2-50+2*8',       '27'],
    [ 6.1, 'print 123.0/2-50+2*8',   '27.5'],
    [ 6.2, 'print 123/2.0-50+2*8',   '27.5'],
    [ 6.3, 'print 123.0/2.0-50+2*8', '27.5'],
    [ 6.4, 'print 123/2-50+2.0*8',   '27'],
    [ 6.5, 'print 123/2-50+2*8.0',   '27'],
    [ 6.6, 'print 123/2-50+2.0*8.0', '27'],
    [ 6.7, 'print 123/2-50.0+2*8',   '27'],
    [ 7, 'print 1-2', '-1'],
    [ 8, 'print 1--2', '3'],
    [ 9, 'print 1- +1', '0'],
    [10,   'print 1+ +1',     '2'],
    [10.1, 'print 1.0+ +1',   '2'],
    [10.2, 'print 1+ +1.0',   '2'],
    [10.3, 'print 1.0+ +1.0', '2'],
    [11,   'print ---10',   '-10'],
    [11.1, 'print ---10.0', '-10'],
    [12,   'print 1==1',     'true'],
    [12.1, 'print 1.0==1',   'true'],
    [12.2, 'print 1==1.0',   'true'],
    [12.3, 'print 1.0==1.0', 'true'],
    [13,   'print not 1>3',     'true'],
    [13.1, 'print not 1.0>3',   'true'],
    [13.2, 'print not 1>3.0',   'true'],
    [13.3, 'print not 1.0>3.0', 'true'],
    [13.4, 'print not 3<=1',    'true'],
    [14,   'print 1<2 and 0 or "whatever"',     '0'],
    [14.1, 'print 1.0<2 and 0 or "whatever"',   '0'],
    [14.2, 'print 1<2.0 and 0 or "whatever"',   '0'],
    [14.3, 'print 1.0<2.0 and 0 or "whatever"', '0'],
    [15,   'print 1>2 or true and 42', '42'],
    [16,   'print 1<10/5 and 127*-1+154', '27'],
    [17, 'print "q"*2', 'qq'],
    [18, 'print "q"+"x"', 'qx'],
    [19, 'print "q"+3', 'q3'],
    [20, 'print "q"+3.14', 'q3.14'],
    [21, 'print "q"=="q"', 'true'],
    [22, 'print "q"!="p"', 'true'],
    [23, 'print "p"<"q"',  'true'],
    [24, 'print "q">"p"',  'true'],
    [25,   'print not (false or true)',  'false'],
    [25.1, 'print not (false and true)', 'true'],
    [25.2, 'print "" or 42',       '42'],
    [26, 'var a=100; var b=a-90; print -b', '-10'],
    [27, 'var a=1; var b=2; print a+b',  '3'],
    [28, 'var a=1; eval a=a+1; print a', '2'],
    [29, 'var a=1; print a=a+1',         '2'],
    [30, 'def blk {print TYPE}',         'blk'],
    [31, 'def blk "foo" {print TYPE+"."+NAME}', 'blk.foo'],
    [32, 'var a=1; def blk {print TYPE}',   'blk'],
    [33, 'print 1; def blk {print TYPE}',   '1\nblk'],
    [34, 'var a=1; def blk {print TYPE+a}', 'blk1'],
    [35, 'def blk {var a=1; print TYPE+a}', 'blk1'],
    [36, 'var x=1; def blk {var a=1+x; print TYPE+a}',   'blk2'],
    [37, 'def blk {var a=1; var b=2; print TYPE+(a+b)}', 'blk3'],
    [38, 'var x=5; def blk {var a=1; var b=2; print TYPE+(a+b+x)}', 'blk8'],
    [39, 'def b1{var x=1; def b2 {var a=2; print TYPE+(a+x)}}',     'b23'],
    [40, 'var a=5; def blk {var a=a+1; print TYPE+a}',   'blk6'],
    [41, 'def b1{var a=5; def b2 {var a=a+1; print TYPE+a} print TYPE+a}', 'b26\nb15'],
    [42, 'def b1{var a=5; def b2 {eval a=a+1; print TYPE+a} print TYPE+a}','b26\nb16'],
    [43, 'def b1{ print TYPE; def b2{print TYPE} }; def b3{print TYPE}', 'b1\nb2\nb3'],
    [44, 'var a; var b; eval a=1+(b=2); print a', '3'],
    [45, 'var a; var b; print a=1+(b=2)',         '3'],
    [46, 'def x {42}', ''],
    [47, 'def x {var x=42; x+1; print x}', '42'],
    [48, 'def x {a=1+(b=2); print a}', '3'],
    [49, 'def x {print a=1+(b=2)}',    '3'],
    [50, 'def x {print (a=1)+(b=2)}',  '3'],

    [51, '', '== input ==\n0000    1:1  RET', 'disasm'],
    [52, 'eval nil',
        '== input ==\n0000    1:9  NIL\n0001      |  POP\n0002      |  RET',
        'disasm'],
    [53, 'eval 42',
        "== input ==\n0000    1:8  CONST         0 '42'\n"
                     "0002      |  POP\n0003      |  RET",
        'disasm'],
    [54, 'def b {}',
        "== input ==\n0000    1:8  DEFBLOCK      0 'b'\t   1 ''\n"
                     "0003    1:9  ENDBLOCK\n0004      |  RET",
        'disasm'],

    [55,   '1', '',           "err: expected statement"],
    [55.1, '=1', '',          "err: expected statement"],
    [56,   'print', '',       "err: at end: expected expression"],
    [56.1, 'print print', '', "err: at 'print': expected expression"],
    [56.2, 'print =', '',     "err: at '=': expected expression"],
    [57,   'eval', '',        "err: expected expression"],
    [58,   'eval (1', '',     "err: expected ')'"],
    [59,   'def 1', '',         "err: at '1': expected block type"],
    [60,   'def b {', '',       "err: at end: expected '}'"],
    [61.1, 'def b x {}', '',    "err: at 'x': expected '{'"],
    [61.2, 'def b 0 {}', '',    "err: at '0': expected '{'"],
    [62.1, 'def b "x" {', '',   "err: at end: expected '}'"],
    [62.2, 'def b "x" { z', '', "err: at end: expected '}'"],
    [63.1, 'eval 1 =', '',      "err: at '=': invalid assignment target"],
    [63.2, 'eval "a" = ', '',   "err: at '=': invalid assignment target"],
    [63.3, 'eval false = ', '', "err: at '=': invalid assignment target"],
    [64,   'eval a=42', '',     "err: at '=': invalid assignment target"],
    [65,   'var a; var a', '',  "err: at 'a': variable with this name already present"],
    [66,   'eval a', '',        "err: at 'a': undefined variable"],
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

    fail = False
    for i, prog, exp, *opt in tests:
        cmd2 = cmd.copy()
        if 'disasm' in opt: cmd2.append('--disasm')

        proc = subprocess.run(cmd2, input=prog, text=True, capture_output=True)

        if proc.returncode == 0:
            if (res := proc.stdout.rstrip()) != exp:
                fail = True
                perr(f'#{i} mismatch: have {res!r}, want {exp!r}')

        else:
            msg = proc.stderr.rstrip().removesuffix("\ncombined errors from parse")
            if m := err_match(opt):
                if m == 'err': continue
                if m not in msg:
                    perr(f'#{i} error mismatch: have {msg!r}, want matching {m!r}')
                    fail = True
            else:
                perr(f'#{i} unexpected error: {msg!r}')
                fail = True

    if fail: exit(1)


def q(s): return s.encode('unicode-escape').decode()


part1 = r"""// Code generated by "./test.py generate"; DO NOT EDIT.

package bcl

import (
    "bytes"
    "strings"
    "testing"
)

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
            out := new(bytes.Buffer)
            log := new(bytes.Buffer)

            _, err := Interpret(
                []byte(tc.input),
                OptDisasm(tc.disasm),
                OptOutput(out), OptLogger(log),
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

TARGET = 'testapi_test.go'


def generate():
    with open(TARGET, 'w') as f:
        print(part1, file=f)

        for (i, inp, outp, *opt) in tests:
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
        generate()
    else:
        run_tests()
