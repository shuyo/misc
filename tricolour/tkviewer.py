#!/usr/bin/env python3
"""GUI viewer for TRICOLOUR matches using tkinter."""

import tkinter as tk
import time
from tricolore import (
    tuple2pos, RED, BLUE,
    RandomPlayer, Greedy, match
)

CELL = 60
COLORS = {
    0: 'pink',       # W_RED
    1: 'lightblue',  # W_BLUE
    2: 'red',        # RED
    3: 'blue',       # BLUE
    4: 'white',      # BLANK
}

class BoardGUI:
    def __init__(self):
        self.root = tk.Tk()
        self.root.title("TRICOLOUR")
        self.canvas = tk.Canvas(self.root, width=CELL*6, height=CELL*6)
        self.canvas.pack()
        self.cells = []
        for y in range(6):
            row = []
            for x in range(6):
                x1, y1 = x*CELL, y*CELL
                rect = self.canvas.create_rectangle(
                    x1, y1, x1+CELL, y1+CELL,
                    outline='black', fill='white'
                )
                row.append(rect)
            self.cells.append(row)
        self.root.update()

    def update(self, board):
        for y in range(6):
            for x in range(6):
                pos = tuple2pos((y, x))
                color = COLORS.get(board[pos], 'white')
                self.canvas.itemconfig(self.cells[y][x], fill=color)
        self.root.update()


def main():
    gui = BoardGUI()
    players = (
        (RED, "RED", RandomPlayer("RED")),
        (BLUE, "BLUE", Greedy("BLUE")),
    )
    match(players, output=False, display=gui.update, delay=0.3)
    gui.root.mainloop()


if __name__ == '__main__':
    main()
