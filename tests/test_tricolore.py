import os
import sys

sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'tricolour'))
import tricolore


def test_initboard_score():
    board = tricolore.initboard()
    (red, blue), blank = tricolore.score(board)
    assert red == 2
    assert blue == 2
    assert blank == 32


def test_availableplaces_start():
    board = tricolore.initboard()
    for stone in (tricolore.RED, tricolore.BLUE):
        stones, whites = tricolore.availableplaces(board, stone)
        assert len(stones) + len(whites) > 0
