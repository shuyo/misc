#!/usr/bin/env python

import random

koma = [
    "OR",       # white (red)
    "OB",       # white (blue)
    "RD",       # red
    "BL",       # blue
    "  ",       # blank
    "XX"]       # gurd

W_RED = 0
W_BLUE= 1
RED   = 2
BLUE  = 3
BLANK = 4
GUARD = 5

directions = [-8,-7,-6,-1,1,6,7,8]

def initboard():
    board = [BLANK for i in xrange(57)]
    for i in xrange(7):
        board[i] = board[50+i] = GUARD
        board[i*7] = GUARD

    board[3*7+3]=board[4*7+4]=RED
    board[3*7+4]=board[4*7+3]=BLUE

    return board

def printboard(board):
    line = ("+--"*6)+"+"
    print line
    for i in xrange(6):
        print "|"+("|".join(koma[x] for x in board[i*7+8:i*7+14]))+"|"
        print line

def isAvailable(board, stone, pos):
    for d in directions:
        i = pos + d
        x = board[i]
        if x >= BLANK or x == stone: continue
        while True:
            i += d
            x = board[i]
            if x == stone: return True
            if x >= BLANK: break
    return False

def isAvailableW(board, pos):
    for d in directions:
        i = pos + d
        x = board[i]
        #if pos==19: print pos, d, i, x
        if x != RED and x != BLUE: continue
        while True:
            i += d
            x = board[i]
            #if pos==19: print pos, d, i, x
            if x <= W_BLUE: return True
            if x >= BLANK: break
    return False

def availableplaces(board, stone):
    stones = []
    whites = []
    for pos in xrange(8, 48):
        if board[pos] != BLANK: continue
        if isAvailable(board, stone, pos):
            stones.append(pos)
        if isAvailableW(board, pos):
            whites.append(pos)
    return stones, whites

def putstone(board, stone, pos):
    if board[pos] != BLANK: raise Exception("cannot put at non-blank position")
    turned = False
    for d in directions:
        i = pos + d
        x = board[i]
        if x >= BLANK or x == stone: continue
        while True:
            i += d
            x = board[i]
            if x == stone:
                i -= d
                while i != pos:
                    board[i] ^= 2
                    turned = True
                    i -= d
                break
            if x >= BLANK: break
    if not turned: raise Exception("turn no discs at the position")
    board[pos] = stone

def putstoneW(board, stone, pos):
    if board[pos] != BLANK: raise Exception("cannot put at non-blank position")
    turned = False
    for d in directions:
        i = pos + d
        x = board[i]
        if x >= BLANK or x <= W_BLUE: continue
        while True:
            i += d
            x = board[i]
            if x <= W_BLUE:
                i -= d
                while i != pos:
                    board[i] ^= 2
                    turned = True
                    i -= d
                break
            if x >= BLANK: break
    if not turned: raise Exception("cannot turn any discs at the position")
    board[pos] = stone ^ 2

def score(board):
    red = blue = 0
    for pos in xrange(8, 48):
        x = board[pos]
        if x == RED: red += 1
        if x == BLUE: blue += 1
    return red, blue


# ----

if __name__ == '__main__':
    board = initboard()
    printboard(board)

    side = RED
    while True:
        SIDE = ("RED" if side == RED else "BLUE")

        stones, whites = availableplaces(board, side)
        #print stones, whites
        list = [(x, 0) for x in stones] + [(x, 1) for x in whites]
        if len(list)==0:
            print SIDE + " passes"

        pos, col = random.choice(list)
        if col == 0:
            print "%s puts %s at %d" % (SIDE, SIDE, pos)
            putstone(board, side, pos)
        else:
            print "%s puts WHITE at %d" % (SIDE, pos)
            putstoneW(board, side, pos)

        side = RED + BLUE - side

        sc = score(board)
        printboard(board)
        print "score:", sc

        if sc[0]==0:
            if sc[1]==0:
                print "draw!"
                break
            print "BLUE won!"
            break
        if sc[1]==0:
            print "RED won!"
            break
