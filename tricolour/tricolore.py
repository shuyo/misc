#!/usr/bin/env python

# Game Matcher and some Clients for Boardgame "TRICOLOUR"

# This code is available under the MIT License.
# (c)2013 Nakatani Shuyo / Cybozu Labs Inc.

import sys, random

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

def pos2tuple(pos):
    return (pos / 7 - 1, pos % 7 - 1)

def tuple2pos(t):
    y, x = t
    return y * 7 + x + 8

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
        if x != RED and x != BLUE: continue
        while True:
            i += d
            x = board[i]
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

def score(board, k1=RED, k2=BLUE):
    red = blue = blank = 0
    for pos in xrange(8, 48):
        x = board[pos]
        if x == k1: red += 1
        elif x == k2: blue += 1
        elif x == BLANK: blank += 1
    return (red, blue), blank

class PlayerBase(object):
    name = "*dummy*"
    def __init__(self, side):
        self.board = initboard()
        self.myside = RED if side == "RED" else BLUE
        self.opponent = RED + BLUE - self.myside
        self.mycolor = side

    def move(self, pos, color):
        pos = tuple2pos(pos)
        if color == "WHITE":
            putstoneW(self.board, self.opponent, pos)
        else:
            putstone(self.board, self.opponent, pos)

    def nextmove(self):
        pass

    def move_return(self, pos, col):
        if col:
            putstone(self.board, self.myside, pos)
            return "MOVE", pos2tuple(pos), self.mycolor
        else:
            putstoneW(self.board, self.myside, pos)
            return "MOVE", pos2tuple(pos), "WHITE"

class RandomPlayer(PlayerBase):
    name = "Random"
    def nextmove(self):
        stones, whites = availableplaces(self.board, self.myside)

        list = [(x, True) for x in stones] + [(x, False) for x in whites]
        if len(list)==0:
            return "PASS", None, None
        return self.move_return(*random.choice(list))

class RandomPlayer2(PlayerBase):
    name = "Random(not leave 1)"
    def nextmove(self):
        count = 0
        for pos in xrange(8, 48):
            if self.board[pos] == self.myside: count += 1

        stones, whites = availableplaces(self.board, self.myside)

        list = [(x, True) for x in stones]
        if len(list)==0 or count > 1: list += [(x, False) for x in whites]
        if len(list)==0:
            return "PASS", None, None
        return self.move_return(*random.choice(list))

class RandomPlayer3(PlayerBase):
    name = "Random(weight to color)"
    def nextmove(self):
        stones, whites = availableplaces(self.board, self.myside)

        list = [(x, True) for x in stones] + [(x, True) for x in stones] + [(x, False) for x in whites]
        if len(list)==0:
            return "PASS", None, None
        return self.move_return(*random.choice(list))

class Greedy(PlayerBase):
    name = "Greedy"
    def nextmove(self):
        stones, whites = availableplaces(self.board, self.myside)
        if len(stones)==0 and len(whites)==0:
            return "PASS", None, None
        best = -sys.maxint, None, None
        for pos in stones:
            board = self.board[:]
            putstone(board, self.myside, pos)
            sc, bl = score(board, self.myside, self.opponent)
            s = sc[0]-sc[1]
            if sc[1]==0: s+=999
            if s > best[0]:
                best = (s, pos, True)
        for pos in whites:
            board = self.board[:]
            putstoneW(board, self.myside, pos)
            sc, bl = score(board, self.myside, self.opponent)
            s = sc[0]-sc[1]
            if sc[1]==0: s+=999
            if s > best[0]:
                best = (s, pos, False)
        return self.move_return(best[1], best[2])

class MinMax(PlayerBase):
    name = "MinMax"
    def nextmove(self):
        stones, whites = availableplaces(self.board, self.myside)
        if len(stones)==0 and len(whites)==0:
            return "PASS", None, None
        best = -sys.maxint, None, None
        for pos in stones:
            board = self.board[:]
            putstone(board, self.myside, pos)
            sc, bl = score(board, self.myside, self.opponent)
            if sc[1]==0: return self.move_return(pos, True)
            worst = self.estimateOpponent(board)
            if worst > best[0]:
                best = (worst, pos, True)
        for pos in whites:
            board = self.board[:]
            putstoneW(board, self.myside, pos)
            sc, bl = score(board, self.myside, self.opponent)
            if sc[1]==0: return self.move_return(pos, False)
            worst = self.estimateOpponent(board)
            if worst > best[0]:
                best = (worst, pos, False)
        return self.move_return(best[1], best[2])

    def estimateOpponent(self, next_board):
        stones, whites = availableplaces(next_board, self.opponent)
        worst = sys.maxint
        for pos in stones:
            board = next_board[:]
            putstone(board, self.opponent, pos)
            sc, bl = score(board, self.myside, self.opponent)
            s = sc[0] - sc[1]
            if sc[0]==0: s-=999
            if s < worst:
                worst = s
        for pos in whites:
            board = next_board[:]
            putstoneW(board, self.opponent, pos)
            sc, bl = score(board, self.myside, self.opponent)
            s = sc[0]-sc[1]
            if sc[0]==0: s-=999
            if s < worst:
                worst = s
        return worst



def match(players, output=None):
    board = initboard()
    if output==True: printboard(board)

    turn = 0
    try:
        passed = 0
        while True:
            side, SIDE, player = players[turn]

            command, pos, col = player.nextmove()

            turn = 1 - turn

            if command == "PASS":
                if output==True: print SIDE + " passes"
                passed += 1
            else:
                if output==True: print "%s puts %s at (%d,%d)" % (SIDE, col, pos[0], pos[1])
                addr = tuple2pos(pos)
                passed = 0
                if col == "WHITE":
                    putstoneW(board, side, addr)
                else:
                    putstone(board, side, addr)
                players[turn][2].move(pos, col)

            sc, bl = score(board)
            if output==True:
                printboard(board)
                print "score:", sc

            if bl == 0 or passed >= 2 or sc[0] * sc[1] == 0:
                if output!=None:
                    if sc[0] > sc[1]:
                        print "RED won!", sc
                    elif sc[0] < sc[1]:
                        print "BLUE won!", sc
                    else:
                        print "draw!"
                return sc
    except Exception, e:
        print "\n[Threw Exception when %s puts %s at (%d,%d) probably]" % (SIDE, col, pos[0], pos[1])
        printboard(board)
        raise

def statistics(player1, player2, N=100, output=None):
    red = blue = draw = 0
    for n in xrange(N):
        players = (RED, "RED", player1("RED")), (BLUE, "BLUE", player2("BLUE"))
        sc = match(players, output)
        if sc[0] > sc[1]:
            red += 1
        elif sc[0] < sc[1]:
            blue += 1
        else:
            draw += 1
        if output: print "(red, blue, draw) =", (red, blue, draw)
    return (red, blue, draw)

if __name__ == '__main__':
    playerlist = [RandomPlayer, RandomPlayer2, RandomPlayer3, Greedy, MinMax]
    L = len(playerlist)
    for i in xrange(L):
        player1 = playerlist[i]
        for j in xrange(L):
            player2 = playerlist[j]
            sc = statistics(player1, player2, 200)
            z = sc[0]+sc[1]
            print "%s vs. %s : %s %.1f" % (player1.name, player2.name, sc, sc[0] / float(sc[0]+sc[1]) * 100 if z > 0 else 0)

