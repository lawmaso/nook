from dataclasses import dataclass


@dataclass
class Palette:
    bg:       str
    surface:  str
    surface2: str
    border:   str
    accent:   str
    accent2:  str
    text:     str
    text_dim: str
    green:    str
    red:      str
    gold_row: str


@dataclass
class AnimalCrossing(Palette):
    bg:       str = "#F6F1E7"  # warm switch dock cream
    surface:  str = "#ECE3D3"  # soft island sand
    surface2: str = "#DDD0BB"  # deeper beige plastic
    border:   str = "#B7A78F"  # warm taupe outline
    accent:   str = "#5DAFBF"  # ocean cyan
    accent2:  str = "#57C299"  # mint green
    text:     str = "#3B3126"  # warm cocoa brown
    text_dim: str = "#756857"  # muted driftwood text
    green:    str = "#7CC7A2"  # island leaf green
    red:      str = "#EC6666"  # soft coral accent
    gold_row: str = "#F0C57A"  # sandy parchment highlight
