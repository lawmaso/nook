import tkinter as tk
from dataclasses import dataclass, field


@dataclass
class TreeviewColumns:
    definitions: dict = field(default_factory=lambda: {
        #  key:     (header, width, anchor, stretch)
        "rank":     ("#",          48,  tk.CENTER, False),
        "user":     ("Username",   100, tk.W,      True),
        "points":   ("Total Pts",  80,  tk.W,      False),
        "solved":   ("Solved",     95,  tk.W,      False),
        "avg_time": ("Avg Time",   76,  tk.W,      False),
        "acc":      ("Acceptance", 76,  tk.W,      False)
    })
