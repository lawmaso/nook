from __future__ import annotations
from typing import TYPE_CHECKING
import tkinter as tk


if TYPE_CHECKING:  # prevents circular import
    from apps.desktop.ui.app import Nook


def get_root(widget: tk.Widget) -> Nook:
    return widget.winfo_toplevel()  # type: ignore[return-value]
