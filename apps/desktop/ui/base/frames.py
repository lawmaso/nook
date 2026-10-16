import tkinter as tk
from apps.desktop.ui.utils.root import get_root


class NookFrame(tk.Frame):
    def __init__(self, parent: tk.Widget, **kwargs):
        t = get_root(parent).theme

        kwargs.setdefault("bg", t.palette.bg)

        super().__init__(parent, **kwargs)


class NookSurfaceFrame(tk.Frame):
    def __init__(self, parent: tk.Widget, **kwargs):
        t = get_root(parent).theme

        kwargs.setdefault("bg", t.palette.surface)

        super().__init__(parent, **kwargs)
