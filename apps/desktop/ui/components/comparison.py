import datetime as dt
from enum import Enum
import tkinter as tk
import tkinter.ttk as ttk
from zoneinfo import ZoneInfo

import apps.desktop.ui.config as config
import apps.desktop.ui.utils.formatters as fmt
from apps.desktop.core.analysis import point_gap, puzzle_gap, time_gap
from apps.desktop.core.models import User
from apps.desktop.ui.base.frames import NookFrame
from apps.desktop.ui.base.labels import *
from apps.desktop.ui.utils.root import get_root


class _NookSelector(NookFrame):
    _label:    NookLabel
    _variable: tk.StringVar
    _combobox: ttk.Combobox
    _on_change: callable

    def __init__(self, parent: tk.Widget, label: str, on_change: callable):
        t = get_root(parent).theme
        super().__init__(parent)

        self._on_change = on_change

        NookLabel(self, text=label).pack(anchor="w")
        self._variable = tk.StringVar()
        self._combobox = ttk.Combobox(self,
            textvariable=self._variable,
            state="readonly",
            font=t.fonts.ui(),
            style="nook.TCombobox"                       
        )

        # handles bug when up/down buttons are used and selection length doesn't change
        self._combobox.bind("<<ComboboxSelected>>", lambda e: self._combobox.selection_clear())
        self._variable.trace_add("write", lambda *_: self._combobox.selection_clear())

        ttk.Button(self,
            text=config.BUTTON_UP,
            command=self._increase_current,
            style="nook.TButton"
        ).pack(side=tk.RIGHT)

        ttk.Button(self, 
            text=config.BUTTON_DOWN,
            command=self._decrease_current,
            style="nook.TButton"
        ).pack(side=tk.RIGHT)

        t.style_combobox(self._combobox)
        self._combobox.pack(fill=tk.X)
        self.pack(fill=tk.X)

    def set_choices(self, choices: list[str]):
        self._combobox["values"] = choices

    def get_current(self) -> int:
        return self._combobox.current()
    
    def set_current(self, index: int):
        self._combobox.current(index)

    def _increase_current(self):
        max_index = len(self._combobox["values"]) - 1
        curr_index = self.get_current()

        # can only increase if index is not the last one
        if curr_index < max_index:
            self.set_current(min(max_index, self.get_current() + 1))
            self._on_change()

    def _decrease_current(self):
        curr_index = self.get_current()

        # can only decrease if index is not 0th one
        if 0 < curr_index:
            self.set_current(max(0, self.get_current() - 1))
            self._on_change()


class _NookSelection(NookFrame):
    _selector_a: _NookSelector
    _selector_b: _NookSelector

    def __init__(self, parent: tk.Widget, on_change: callable):
        super().__init__(parent)
        self._on_change = on_change
        self._build(on_change)

    def _build(self, on_change: callable):
        t = get_root(self).theme

        NookHeaderLabel(self, text=config.COMPARISON_TEXT).pack(anchor="w")
        self._selector_a = _NookSelector(self, label=config.SELECTION_U1, on_change=on_change)
        self._selector_b = _NookSelector(self, label=config.SELECTION_U2, on_change=on_change)

    def set_users(self, users: list[User]):
        choices = [f"#{u.rank} {u.username}" for u in users[1:]]
        self._selector_a.set_choices(choices)
        self._selector_b.set_choices(choices)

        self._selector_a.set_current(0)
        self._selector_b.set_current(1)

    def get_selections(self) -> tuple[int, int]:
        return (
            self._selector_a.get_current(),
            self._selector_b.get_current()
        )


class _StatDirection(Enum):
    HIGHER = "higher"
    LOWER = "lower"


class _NookAnalysis(NookFrame):
    last_active_labels: list[NookLabel]
    _tick_id: str | None = None

    def __init__(self, parent: tk.Widget):
        super().__init__(parent)
        self.pack(fill=tk.BOTH, expand=True)
        self.last_active_labels = []

    def render(self, u1: User, u2: User):
        t = get_root(self).theme
        self._clear()

        NookHeaderLabel(self, text=config.ANALYIS_TEXT).pack(
            anchor="w",
            pady=t.spacing.section_py
        )

        stats = [
            (point_gap(u1, u2), fmt.point_gap, config.STAT_POINT_GAP),
            (puzzle_gap(u1, u2), fmt.puzzle_gap, config.STAT_PUZZLE_GAP),
            (time_gap(u1, u2), fmt.time_gap, config.STAT_TIME_GAP)
        ]

        for stat, formatter, description in stats:
            NookStatLabel(self, text=formatter(stat)).pack()
            NookLabel(self, text=description).pack()

        self._render_comparison_table(u1, u2)
        self._render_last_active_labels(u1, u2)

    def _render_comparison_table(self, u1: User, u2: User):
        t = get_root(self).theme
        green, red = t.palette.green, t.palette.red

        table = NookFrame(self)
        table.pack(fill=tk.X, pady=t.spacing.top_py)

        for col in range(3):
            table.columnconfigure(col, weight=1, minsize=50)

        NookLabel(table, text="Category", anchor="w", font=t.fonts.title()) \
            .grid(row=0, column=0, sticky="w")
        NookLabel(table, text=f"#{u1.rank}", fg=green, anchor="w", font=t.fonts.title()) \
            .grid(row=0, column=1, sticky="w")
        NookLabel(table, text=f"#{u2.rank}", fg=red, anchor="w", font=t.fonts.title()) \
            .grid(row=0, column=2, sticky="w")

        stats = [
            ("Total pts",  u1.total_points,    u2.total_points,    fmt.format_large, _StatDirection.HIGHER),
            ("Solved",     u1.solved,          u2.solved,          fmt.format_large, _StatDirection.HIGHER),
            ("Avg pts",    u1.average_points,  u2.average_points,  fmt.format_large, _StatDirection.HIGHER),
            ("Avg time",   u1.average_time,    u2.average_time,    fmt.seconds,      _StatDirection.LOWER),
            ("Acceptance", u1.acceptance_rate, u2.acceptance_rate, fmt.percent,      _StatDirection.HIGHER)
        ]

        for row, (label, va, vb, formatter, better) in enumerate(stats, start=1):
            ca, cb = self._get_stat_colors(va, vb, better)
            NookLabel(table, text=label).grid(row=row, column=0, sticky="w")
            NookLabel(table, text=formatter(va), fg=ca).grid(row=row, column=1, sticky="w")
            NookLabel(table, text=formatter(vb), fg=cb).grid(row=row, column=2, sticky="w")

    def _get_stat_colors(self, va, vb, better: _StatDirection) -> tuple[str, str]:
        t = get_root(self).theme

        neutral = t.palette.text_dim
        green   = t.palette.green
        red     = t.palette.red

        equivalent   = (neutral, neutral)
        better_worse = (green, red)
        worse_better = reversed(better_worse)

        if va == vb:
            return equivalent
        elif better == _StatDirection.LOWER:
            if va < vb:
                return better_worse
            return worse_better
        elif better == _StatDirection.HIGHER:
            if va > vb:
                return better_worse
            return worse_better

    def _render_last_active_labels(self, u1: User, u2: User):
        t = get_root(self).theme

        self.last_active_labels.clear()
        for u, col in [(u1, t.palette.green), (u2, t.palette.red)]:
            NookLabel(
                self,
                text=f"[#{u.rank}] {u.username}",
                fg=col,
                font=t.fonts.title()
            ).pack(anchor="w", pady=(t.spacing.top_py if u == u1 else 0))
            
            label = NookLabel(self, text="")
            label.pack(anchor="w")
            label._anchor_time = u.last_active

            self.last_active_labels.append(label)

        self._tick()

    def _tick(self):
        if not self.last_active_labels:
            return
        
        now = dt.datetime.now(tz=ZoneInfo("America/New_York"))

        for label in self.last_active_labels:
            delta_seconds = int((now - label._anchor_time).total_seconds())
            delta_str = fmt.time_gap(delta_seconds)
            label.config(text=f"Last active: {delta_str} ago")

        self._tick_id = self.after(1000, self._tick)

    def _clear(self):
        if tick_id := getattr(self, "_tick_id", None):
            self.after_cancel(tick_id)
            self._tick_id = None

        self.last_active_labels.clear()
        for widget in self.winfo_children():
            widget.destroy()

    def clear(self):
        self._clear()


class NookComparison(NookFrame):
    _selection:  _NookSelection
    _analysis:   _NookAnalysis
    _on_compare: callable

    def __init__(self, parent: tk.Widget, on_compare: callable):
        t = get_root(parent).theme
        super().__init__(parent, width=t.spacing.compare_width)
        self.pack_propagate(flag=False)
        self._on_compare = on_compare
        self._build()
        self.pack(
            side=tk.RIGHT,
            fill=tk.Y,
            padx=t.spacing.section_px,
            pady=t.spacing.section_py  
        )

    def _build(self):
        t = get_root(self).theme

        self._selection = _NookSelection(self, on_change=self._on_compare)
        self._selection.pack(fill=tk.X)

        compare = NookActionLabel(self, text=config.BUTTON_COMPARE)
        compare.bind("<Button-1>", lambda e: self._on_compare())
        compare.pack(fill=tk.X, pady=t.spacing.top_py)

        self._analysis = _NookAnalysis(self)

    def set_users(self, users: list[User]):
        self._selection.set_users(users)
        self._analysis.clear()

    def get_selections(self) -> tuple[int, int]:
        return self._selection.get_selections()
    
    def render(self, u1: User, u2: User):
        self._analysis.render(u1, u2)
