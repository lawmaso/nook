import tkinter as tk

import apps.desktop.ui.config as config
from apps.desktop.ui.base.frames import NookSurfaceFrame
from apps.desktop.ui.base.labels import NookActionLabel, NookHeaderLabel, NookLabel
from apps.desktop.ui.utils.root import get_root


class NookHeader(NookSurfaceFrame):
    _refresh:    NookActionLabel
    _status:     NookLabel
    _on_refresh: callable

    def __init__(self, parent: tk.Widget, on_refresh: callable):
        t = get_root(parent).theme
        self._on_refresh = on_refresh

        super().__init__(parent, pady=t.spacing.header_py)
        self._build()

    def _build(self):
        t = get_root(self).theme

        self.pack(fill=tk.X)

        NookHeaderLabel(
            self,
            text=config.APP_HEADER,
            bg=t.palette.surface,
            font=t.fonts.heading()
        ).pack(side=tk.LEFT, padx=t.spacing.header_px)

        refresh = NookActionLabel(self, text=config.BUTTON_REFRESH)
        refresh.bind("<Button-1>", lambda e: self._on_refresh())
        refresh.pack(side=tk.RIGHT, padx=t.spacing.btn_px)
        self._refresh = refresh

        status = NookLabel(self, text=config.STATUS_LOADING, bg=t.palette.surface)
        status.pack(side=tk.RIGHT, padx=t.spacing.status_px)
        self._status = status

    def set_status(self, text: str, error: bool = False):
        t = get_root(self).theme

        self._status.config(text=text, fg=(
            t.palette.red if error else t.palette.text_dim)
        )
