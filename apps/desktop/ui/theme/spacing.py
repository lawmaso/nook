from dataclasses import dataclass


@dataclass
class Spacing:
    header_px:     int = 10
    header_py:     int = 16
    section_px:    int = (12, 12)
    section_py:    int = (12, 12)
    compare_width: int = 290
    btn_px:        int = 16
    btn_py:        int = 8
    status_px:     int = 16
    status_py:     int = 8
    left_px:       tuple = (12, 0)
    right_px:      tuple = (0, 12)
    top_py:        tuple = (12, 0)
    bottom_py:     tuple = (0, 12)
