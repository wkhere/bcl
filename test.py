#!/usr/bin/env python3

import subprocess
from sys import stderr, exit

tests = [
    [ 0, 'eval "expr that is discarded"', ''],
    [ 0.1, 'var a; print not a', 'true'],
    [ 0.2, 'var a; print a==nil', 'true'],
    [ 0.3, 'var a=1; eval a=nil; print a==nil', 'true'],
    [ 1, 'print 1+1', '2'],
    [ 2, 'print 1+2.14', '3.14'],
    [ 3, 'print 123/2-50+2*8', '27'],
    [ 4, 'print 1-2', '-1'],
    [ 5, 'print 1--2', '3'],
    [ 5.1, 'print 1- +1', '0'],
    [ 5.2, 'print 1+ +1', '2'],
    [ 6, 'print ---10', '-10'],
    [ 7, 'print not 1>3', 'true'],
    [ 8, 'print 1<2 and 0 or "whatever"', '0'],
    [ 9, 'print 1>2 or true and 42', '42'],
    [10, 'print 1<10/5 and 127*-1+154', '27'],
    [11, 'print "q"*2', 'qq'],
    [12, 'print "q"+"x"', 'qx'],
    [13, 'print "q"+3', 'q3'],
    [14, 'print "q"+3.14', 'q3.14'],
    [15, 'var a=100; var b=a-90; print -b', '-10'],
    [16, 'var a=1; var b=2; print a+b',  '3'],
    [17, 'var a=1; eval a=a+1; print a', '2'],
    [18, 'var a=1; print a=a+1',         '2'],
    [19, 'def blk {print TYPE}',         'blk'],
    [20, 'def blk "foo" {print TYPE+"."+NAME}', 'blk.foo'],
    [21, 'var a=1; def blk {print TYPE}',   'blk'],
    [22, 'print 1; def blk {print TYPE}',   '1\nblk'],
    [23, 'var a=1; def blk {print TYPE+a}', 'blk1'],
    [24, 'def blk {var a=1; print TYPE+a}', 'blk1'],
    [25, 'var x=1; def blk {var a=1+x; print TYPE+a}',   'blk2'],
    [26, 'def blk {var a=1; var b=2; print TYPE+(a+b)}', 'blk3'],
    [27, 'var x=5; def blk {var a=1; var b=2; print TYPE+(a+b+x)}', 'blk8'],
    [28, 'def b1{var x=1; def b2 {var a=2; print TYPE+(a+x)}}',     'b23'],
    [29, 'var a=5; def blk {var a=a+1; print TYPE+a}',   'blk6'],
    [30, 'def b1{var a=5; def b2 {var a=a+1; print TYPE+a} print TYPE+a}', 'b26\nb15'],
    [31, 'def b1{var a=5; def b2 {eval a=a+1; print TYPE+a} print TYPE+a}','b26\nb16'],
    [32, 'def b1{ print TYPE; def b2{print TYPE} }; def b3{print TYPE}', 'b1\nb2\nb3'],
    [33, 'var a; var b; eval a=1+(b=2); print a', '3'],
    [34, 'var a; var b; print a=1+(b=2)',         '3'],
    [35, 'def x {42}', ''],
    [36, 'def x {var x=42; x+1; print x}', '42'],
    [37, 'def x {a=1+(b=2); print a}', '3'],
    [38, 'def x {print a=1+(b=2)}',    '3'],
]

def perr(*args):
    print(*args, file=stderr)

def run_tests():
    cmd = './bcl'

    had_err = False
    for i, prog, exp in tests:
        try:
            res = subprocess.check_output(cmd.split(), input=prog, text=True)
            res = res.strip()
            if res != exp:
                had_err = True
                perr(f'#{i} mismatch: have {res!r}, want {exp!r}')

        except subprocess.CalledProcessError:
            perr(f'#{i} {cmd} error')
            had_err = True

    if had_err: exit(1)

if __name__ == '__main__': run_tests()
