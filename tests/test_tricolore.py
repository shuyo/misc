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


def test_pos_tuple_roundtrip():
    for y in range(6):
        for x in range(6):
            pos = tricolore.tuple2pos((y, x))
            assert tricolore.pos2tuple(pos) == (y, x)


def test_putstone_flip():
    board = tricolore.initboard()
    move = tricolore.tuple2pos((1, 3))
    tricolore.putstone(board, tricolore.RED, move)
    (red, blue), blank = tricolore.score(board)
    assert red == 3
    assert blue == 1
    assert blank == 31


def test_putstoneW_flips_to_white():
    board = tricolore.initboard()
    board[tricolore.tuple2pos((2, 2))] = tricolore.RED
    board[tricolore.tuple2pos((2, 3))] = tricolore.RED
    board[tricolore.tuple2pos((2, 4))] = tricolore.W_RED
    move = tricolore.tuple2pos((2, 1))
    tricolore.putstoneW(board, tricolore.BLUE, move)
    assert board[move] == tricolore.W_BLUE
    assert board[tricolore.tuple2pos((2, 2))] == tricolore.W_RED
    assert board[tricolore.tuple2pos((2, 3))] == tricolore.W_RED
    (red, blue), blank = tricolore.score(board)
    assert red == 1
    assert blue == 1
    assert blank == 30


def test_match_display_callback():
    boards = []

    def record(board):
        boards.append(board[:])

    players = (
        (tricolore.RED, "RED", tricolore.RandomPlayer("RED")),
        (tricolore.BLUE, "BLUE", tricolore.RandomPlayer("BLUE")),
    )
    tricolore.match(players, output=False, display=record, delay=0)
    assert len(boards) > 0
