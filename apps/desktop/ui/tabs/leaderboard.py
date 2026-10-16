import tkinter as tk
import tkinter.ttk as ttk
import threading
from dataclasses import dataclass
from typing import Callable

import apps.desktop.ui.config as config
from apps.desktop.core.analysis import *
from apps.desktop.core.models import User
from apps.desktop.ui.base.frames import NookFrame
from apps.desktop.ui.components.comparison import NookComparison
from apps.desktop.ui.components.header import NookHeader
from apps.desktop.ui.components.leaderboard import NookLeaderboard


@dataclass
class NookTabState:
    users:      list[User] | None = None
    loading:    bool              = False
    error:      str        | None = None


class NookLeaderboardTab(NookFrame):
    header:          NookHeader
    leaderboard:     NookLeaderboard
    comparison:      NookComparison
    state:           NookTabState
    _on_load:        Callable[[], list[User]]

    def __init__(self, notebook: ttk.Notebook, on_load: Callable[[], list[User]]):
        super().__init__(notebook)
        self._on_load = on_load
        self.state = NookTabState()
        self._build()
        self._refresh_leaderboard()

    def _build(self):
        self.header = NookHeader(self, on_refresh=self._refresh_leaderboard)
        self.comparison = NookComparison(self, on_compare=self._compare_users)
        self.leaderboard = NookLeaderboard(self)

        # IMPORTANT: build comparison first due to fixed width
        #   - without this ordering, pack_propagate=False has no effect
        #   - allows leaderboard to be collapsed, while comparison is not

    def _refresh_leaderboard(self):
        if self.state.loading:
            return

        self.state.loading = True
        self.header.set_status(text=config.STATUS_LOADING)
        threading.Thread(target=self._load_leaderboard, daemon=True).start()

    def _load_leaderboard(self):
        try:
            users = self._on_load()
            self.after(0, self._on_load_success, users)
        except Exception as e:
            self.after(0, self._on_load_error, str(e))

    def _on_load_success(self, users: list[User]):
        # update tab state
        self.state.loading = False
        self.state.users = users

        # propagate users to sub-components
        self.header.set_status(config.STATUS_SUCCESS.format(n=len(self.state.users) - 1))
        self.leaderboard.load(self.state.users)
        self.comparison.set_users(self.state.users)

        # compare top users
        self._compare_users()

    def _on_load_error(self, message: str):
        self.state.loading = False
        self.state.error = message
        self.header.set_status(text=message, error=True)

    def _compare_users(self):
        i1, i2 = self.comparison.get_selections()
        u1, u2 = self.state.users[i1 + 1], self.state.users[i2 + 1]

        if not u1 or not u2:
            return
        
        u1, u2 = (u1, u2) if u1.rank < u2.rank else (u2, u1)
        self.comparison.render(u1, u2)
