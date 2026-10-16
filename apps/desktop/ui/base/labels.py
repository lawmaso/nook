import tkinter as tk
from apps.desktop.ui.utils.root import get_root


class NookLabel(tk.Label):
    def __init__(self, parent, **kwargs):
        t = get_root(parent).theme

        kwargs.setdefault("bg", t.palette.bg)
        kwargs.setdefault("fg", t.palette.text_dim)
        kwargs.setdefault("font", t.fonts.small())

        super().__init__(parent, **kwargs)


class NookHeaderLabel(tk.Label):
    def __init__(self, parent: tk.Widget, **kwargs):
        t = get_root(parent).theme

        kwargs.setdefault("bg", t.palette.bg)
        kwargs.setdefault("fg", t.palette.text_dim)
        kwargs.setdefault("font", t.fonts.title())

        super().__init__(parent, **kwargs)


class NookStatLabel(tk.Label):
    def __init__(self, parent: tk.Widget, **kwargs):
        t = get_root(parent).theme

        kwargs.setdefault("bg", t.palette.bg)
        kwargs.setdefault("fg", t.palette.accent)
        kwargs.setdefault("font", t.fonts.stat_large())

        super().__init__(parent, **kwargs)


class NookActionLabel(tk.Label):
    def __init__(self, parent: tk.Widget, **kwargs):
        t = get_root(parent).theme

        kwargs.setdefault("bg", t.palette.accent)
        kwargs.setdefault("fg", t.palette.bg)
        kwargs.setdefault("font", t.fonts.title())
        kwargs.setdefault("padx", t.spacing.btn_px)
        kwargs.setdefault("pady", t.spacing.btn_py)
        kwargs.setdefault("cursor", "watch")
        
        super().__init__(parent, **kwargs)
        
        self.bind("<Enter>", lambda e: self.config(bg=t.palette.accent2))
        self.bind("<Leave>", lambda e: self.config(bg=t.palette.accent))
