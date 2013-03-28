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


class Player(object):
    def __init__(self, side):
        pass
    def move(self, pos, color):
        pass
    def nextmove(self):
        pass

class RandomPlayer(Player):
    def __init__(self, side):
        self.board = initboard()
        self.myside = side
        self.opponent = RED + BLUE - side
        self.mycolor = "RED" if side == RED else "BLUE"

    def move(self, pos, color):
        if color == "WHITE":
            putstoneW(self.board, self.opponent, pos)
        else:
            putstone(self.board, self.opponent, pos)

    def nextmove(self):
        stones, whites = availableplaces(self.board, self.myside)

        list = [(x, True) for x in stones] + [(x, False) for x in whites]
        if len(list)==0:
            return "PASS", None, None

        pos, col = random.choice(list)
        if col:
            putstone(self.board, self.myside, pos)
            return "MOVE", pos, self.mycolor
        else:
            putstoneW(self.board, self.myside, pos)
            return "MOVE", pos, "WHITE"


def match(players):
    board = initboard()
    printboard(board)

    turn = 0
    while True:
        side, SIDE, player = players[turn]

        command, pos, col = player.nextmove()

        turn = 1 - turn

        if command == "PASS":
            print SIDE + " passes"
        else:
            print "%s puts %s at (%d,%d)" % (SIDE, col, pos / 7, pos % 7)
            if col == "WHITE":
                putstoneW(board, side, pos)
            else:
                putstone(board, side, pos)
            players[turn][2].move(pos, col)

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

if __name__ == '__main__':
    players = ((RED, "RED", RandomPlayer(RED)), (BLUE, "BLUE", RandomPlayer(BLUE)))
    match(players)

