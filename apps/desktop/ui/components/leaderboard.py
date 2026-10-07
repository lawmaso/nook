import tkinter as tk
import tkinter.ttk as ttk
from core.models import User

import apps.desktop.ui.config as config
from apps.desktop.ui.base.frames import NookFrame
from apps.desktop.ui.base.labels import NookHeaderLabel
from apps.desktop.ui.utils.root import get_root


class NookLeaderboard(NookFrame):
    _tree: ttk.Treeview

    def __init__(self, parent: tk.Widget):
        super().__init__(parent)
        self._build()

    def _build(self):
        t = get_root(self).theme
        self.pack(
            side=tk.RIGHT,
            fill=tk.BOTH,
            expand=True,
            padx=t.spacing.section_px,
            pady=t.spacing.section_py
        )

        NookHeaderLabel(self, text=config.LEADERBOARD_TEXT).pack(anchor="w")

        tree_frame = NookFrame(self, bg=t.palette.border)
        tree_frame.pack(fill=tk.BOTH, expand=True)

        yscroll = ttk.Scrollbar(tree_frame, orient="vertical",
                                style="nook.Vertical.TScrollbar")
        xscroll = ttk.Scrollbar(tree_frame, orient="horizontal",
                                style="nook.Horizontal.TScrollbar")

        tree = ttk.Treeview(tree_frame,
            columns=list(t.columns.definitions.keys()),
            show="headings",
            yscrollcommand=yscroll.set,
            xscrollcommand=xscroll.set,
            selectmode="browse",
            style="nook.Treeview"
        )

        yscroll.config(command=tree.yview)
        xscroll.config(command=tree.xview)

        yscroll.pack(side=tk.RIGHT, fill=tk.Y)
        xscroll.pack(side=tk.BOTTOM, fill=tk.X)
        tree.pack(fill=tk.BOTH, expand=True)

        for col_id, (heading, width, anchor, stretch) \
            in t.columns.definitions.items():
            tree.heading(col_id, text=heading, anchor=anchor)
            tree.column(col_id, width=width, minwidth=width,
                        anchor=anchor, stretch=stretch)

        tree.tag_configure("odd", background=t.palette.surface)
        tree.tag_configure("even", background=t.palette.surface2)
        tree.tag_configure("gold",
            background=t.palette.gold_row,
            foreground=t.palette.accent
        )

        self._tree = tree

    def load(self, users: list[User]):
        self.clear()
        t = get_root(self).theme

        for i, u in enumerate(users):
            if i == 0: continue  # skip sentinel user

            tag = t.row_tag(u.rank, i)
            self._tree.insert("",
                index=tk.END,
                iid=str(u.rank),
                tags=(tag,),
                values=u.row_values
            )

    def clear(self):
        self._tree.delete(*self._tree.get_children())
