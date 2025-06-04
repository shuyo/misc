#!/usr/bin/env python3

# Game Tree Search for Boardgame "TRICOLOUR"

# This code is available under the MIT License.
# (c)2013 Nakatani Shuyo / Cybozu Labs Inc.

import sys
from tricolore import *

MAXDEPTH = int(sys.argv[1]) if len(sys.argv) > 1 else 3

km = [
    "W",       # white (red)
    "W",       # white (blue)
    "R",       # red
    "B"]       # blue

def tokey(side, pos, sc):
    y, x = pos2tuple(pos)
    return "%s%d%d:%d-%d" % (km[side], y, x, sc[0], sc[1])


board = initboard()
#for i in range(6):
#    board[8+i] = board[43+i] = board[8+i*7] = board[13+i*7] = GUARD

tree = dict()
pos = 18
putstone(board, RED, pos)
sc, bl = score(board)
key = tokey(RED, pos, sc)
node=tree[key]=dict()

queue = [(1, board, node)]
while True:
    turn, board, node = queue.pop(0)
    turn += 1
    if turn > MAXDEPTH: break
    side = RED if turn % 2 == 1 else BLUE
    stones, whites = availableplaces(board, side)
    for pos in stones:
        b = board[:]
        putstone(b, side, pos)
        sc, bl = score(b)
        key = tokey(side, pos, sc)
        new_node=node[key]=dict()

        if sc[0] > 0 and sc[1] > 0:
            queue.append((turn, b, new_node))
    for pos in whites:
        b = board[:]
        putstoneW(b, side, pos)
        sc, bl = score(b)
        key = tokey(side^2, pos, sc)
        new_node=node[key]=dict()

        if sc[0] > 0 and sc[1] > 0:
            queue.append((turn, b, new_node))

count = [0] * MAXDEPTH
def printtree(node, depth):
    if len(node)>0: count[depth] += len(node)
    for key in node:
        print("\t" * depth, key)
        printtree(node[key], depth+1)
printtree(tree, 0)
print(count)
