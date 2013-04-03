#!/usr/bin/env python

# Complete Game Tree Analyzer for Boardgame "TRICOLOUR"

# This code is available under the MIT License.
# (c)2013 Nakatani Shuyo / Cybozu Labs Inc.

import sys, re, collections

rm = re.compile(r"^ ?(\t*)([RBW])([0-5])([0-5]):([0-9]+)-([0-9]+)")
SIDE = ("RED", "BLUE")

# load & pruning
tree = dict()
moves = [tree]
maxdepth = 0
#states = collections.defaultdict(int)
with open(sys.argv[1], "rb") as f:
    for s in f:
        m = rm.search(s)
        if m:
            list = m.groups()
            t = len(list[0])
            c = list[1]
            y,x,r,b = (int(z) for z in list[2:])

            side = t % 2 # 0=red, 1=blue
            if t > maxdepth: maxdepth = t
            key = c, y, x, r, b
            #print t, key

            moves = moves[:t+1]
            parent = moves[t]
            parent[key] = node = {"":set(), "parent":parent}
            moves.append(node)

def path1(node, depth):
    """judgement win or lose at leaf"""
    depth += 1
    for key, nextnode in node.iteritems():
        if key == "" or key == "parent": continue
        if len(nextnode) == 2: # at leaf
            c, y, x, r, b = key
            if depth % 2==0: r,b = b,r
            result = "win" if r>b else "lose" if r<b else "draw"
            nextnode[""].add(result)
            if result == "win":
                node[""].add("to lose")
        path1(nextnode, depth)

def path2(node, depth):
    depth += 1
    all_lose = True
    for key, nextnode in node.iteritems():
        if key == "" or key == "parent": continue
        if "lose" not in nextnode[""] and "to lose" not in nextnode[""]:
            all_lose = False
        if len(nextnode)>2: path2(nextnode, depth)
    if all_lose and "" in node:
        node[""].add("to win")
        node["parent"][""].add("to lose")

def printtree(node, depth):
    depth += 1
    for key, nextnode in node.iteritems():
        if key == "" or key == "parent": continue
        print "\t" * depth, "%s%d%d:%d-%d" % key, " ".join(nextnode[""])
        printtree(nextnode, depth)


path1(tree, 0)
for i in xrange(6):
    path2(tree, 0)

printtree(tree, 0)
