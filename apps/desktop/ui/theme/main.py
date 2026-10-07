import tkinter.ttk as ttk
from dataclasses import dataclass, field

from apps.desktop.ui.theme.columns import TreeviewColumns
from apps.desktop.ui.theme.palettes import Palette, AnimalCrossing
from apps.desktop.ui.theme.spacing import Spacing
from apps.desktop.ui.theme.typography import Typography


@dataclass
class Theme:
    style:   ttk.Style | None = None
    palette: Palette          = field(default_factory=Palette)
    fonts:   Typography       = field(default_factory=Typography)
    spacing: Spacing          = field(default_factory=Spacing)
    columns: TreeviewColumns  = field(default_factory=TreeviewColumns)

    def apply_ttk_styles(self, style: ttk.Style):
        self.style = style
        self.style.theme_use("clam")

        self._configure_notebook_styles()
        self._configure_treeview_styles()
        self._configure_combobox_styles()
        self._configure_button_styles()

        self._apply_custom_options()
        self.style.configure("TSeparator", background=self.palette.border)

    def _configure_notebook_styles(self):
        c, f = self.palette, self.fonts

        self.style.configure(
            "nook.TNotebook",
            background=c.bg,
            borderwidth=0,
            tabmargins=[0, 0, 0, 0]
        )
        self.style.configure(
            "nook.TNotebook.Tab",
            background=c.surface,
            foreground=c.text_dim,
            font=f.title(),
            borderwidth=0,
        )
        self.style.map(
            "nook.TNotebook.Tab",
            background=[("selected", c.surface2), ("active", c.surface2)],
            foreground=[("selected", c.accent),   ("active", c.text)],
        )

        # gets rid of focus border on tab selection
        self.style.layout("Tab",
            [("Notebook.tab", {"sticky": "nswe", "children":
                [("Notebook.padding", {"side": "top", "sticky": "nswe", "children":
                    #[("Notebook.focus", {"side": "top", "sticky": "nswe", "children":
                        [("Notebook.label", {"side": "top", "sticky": ""})],
                    #})],
                })],
            })]
        )

    def _configure_treeview_styles(self):
        c, f = self.palette, self.fonts
        tk = self.style.master.tk

        self.style.configure(
            "nook.Treeview",
            background=c.surface,
            foreground=c.text,
            fieldbackground=c.surface,
            rowheight=22,
            font=f.ui(),
            borderwidth=0
        )
        self.style.configure(
            "nook.Treeview.Heading",
            background=c.surface2,
            foreground=c.accent,
            font=f.button(),
            relief="flat"
        )
        self.style.map(
            "nook.Treeview",
            background=[("selected", c.accent2)],
            foreground=[("selected", c.bg)]
        )

        self.style.configure(
            "nook.Vertical.TScrollbar",
            background=c.surface2,
            troughcolor=c.surface,
            arrowcolor=c.accent,
            borderwidth=0
        )

        # handles strange tkinter extension glitch;
        # resets styles when the scrollbar has no scroll area
        self.style.map(
            "nook.Vertical.TScrollbar",
            background=[("disabled", c.surface2), ("active", c.surface2), ("!active", c.surface2)],
            troughcolor=[("disabled", c.surface), ("active", c.surface), ("!active", c.surface)],
            arrowcolor=[("disabled", c.accent), ("active", c.accent), ("!active", c.accent)]
        )

        self.style.configure(
            "nook.Horizontal.TScrollbar",
            background=c.surface2,
            troughcolor=c.surface,
            arrowcolor=c.accent,
            borderwidth=0
        )

        # handles strange tkinter extension glitch;
        # resets styles when the scrollbar has no scroll area
        self.style.map(
            "nook.Horizontal.TScrollbar",
            background=[("disabled", c.surface2), ("active", c.surface2), ("!active", c.surface2)],
            troughcolor=[("disabled", c.surface), ("active", c.surface), ("!active", c.surface)],
            arrowcolor=[("disabled", c.accent), ("active", c.accent), ("!active", c.accent)]
        )

        # tk.call("ttk::style", "configure", "Treeview",
        #     "-bordercolor",  c.surface,
        #     "-lightcolor",   c.surface,
        #     "-darkcolor",    c.surface,
        #     "-relief",       "flat")
        # tk.call("ttk::style", "configure", "Treeview.Heading",
        #         "-bordercolor",  c.surface2,
        #         "-lightcolor",   c.surface2,
        #         "-darkcolor",    c.surface2,
        #         "-relief",       "flat")

    def _configure_combobox_styles(self):
        c, f = self.palette, self.fonts
        tk = self.style.master.tk

        self.style.configure(
            "nook.TCombobox",
            fieldbackground=c.surface2,
            background=c.surface2,
            foreground=c.text,
            selectbackground=c.accent2,
            selectforeground=c.bg,
            arrowcolor=c.accent,
            borderwidth=1,
            relief="flat"
        )
        self.style.map(
            "nook.TCombobox",
            fieldbackground=[("readonly", c.surface2)],
            foreground=[("readonly", c.text)]
        )

        self.style.configure(
            "nook.TCombobox.Vertical.TScrollbar",
            background=c.surface2,
            troughcolor=c.surface,
            arrowcolor=c.accent,
            borderwidth=0
        )

        # handles strange tkinter extension glitch;
        # resets styles when the scrollbar has no scroll area
        self.style.map(
            "nook.TCombobox.Vertical.TScrollbar",
            background=[("disabled", c.surface2), ("active", c.surface2), ("!active", c.surface2)],
            troughcolor=[("disabled", c.surface), ("active", c.surface), ("!active", c.surface)],
            arrowcolor=[("disabled", c.accent), ("active", c.accent), ("!active", c.accent)]
        )

        # style.master.tk.eval("set popdown [ttk::Combobox::PopdownWindow %s]" % cb)

        # tk.call("ttk::style", "configure", "TCombobox",
        #     "-bordercolor",  c.surface2,
        #     "-lightcolor",   c.surface2,
        #     "-darkcolor",    c.surface2,
        #     "-arrowcolor",   c.accent,
        #     "-relief",       "flat")

    def _configure_button_styles(self):
        c, f = self.palette, self.fonts

        self.style.configure(
            "nook.TButton",
            background=c.surface2,
            foreground=c.accent,
            font=f.ui(),
            padding=(8, 1.25),  # (horizontal, vertical) — kills extra vertical padding,
            width=2
        )

        # self.style.map(
        #     "nook.TButton",
        #     background=[("active", c.surface2), ("pressed", c.surface2)],
        #     foreground=[("active", c.accent),   ("pressed", c.accent)],
        # )

        # gets rid of focus border on button click
        self.style.layout("nook.TButton",
            [("Button.border", {"sticky": "nswe", "children":
                [("Button.padding", {"sticky": "nswe", "children":
                    #[("Button.focus", {"sticky": "nswe", "children":
                        [("Button.label", {"sticky": "nswe"})],
                    #})],
                })],
            })]
        )

    def style_combobox(self, cb: ttk.Combobox):
        """
        Attempts to apply custom styles to a combobox
        """

        try:
            self.style.master.tk.eval("set cb [ttk::combobox::PopdownWindow %s]" % cb)
            self.style.master.tk.eval("$cb.f.sb configure -style nook.TCombobox.Vertical.TScrollbar")
        except Exception:
            pass  # the tk version doesn't support these evals

    def _apply_custom_options(self):
        c, f = self.palette, self.fonts

        # overrides combobox default styling
        self.style.master.option_add("*TCombobox*Listbox.background", c.surface)
        self.style.master.option_add("*TCombobox*Listbox.font", f.ui())
        self.style.master.option_add("*TCombobox*Listbox.foreground", c.text)
        self.style.master.option_add("*TCombobox*Listbox.selectBackground", c.accent2)
        self.style.master.option_add("*TCombobox*Listbox.selectForeground", c.bg)

    def row_tag(self, rank: int, index: int) -> str:
        if rank == 1:
            return "gold"
        return "even" if index % 2 == 0 else "odd"


DEFAULT_THEME = Theme(palette=AnimalCrossing)
