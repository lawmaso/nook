import time
import apps.desktop.ui.config as config
from apps.desktop.core.models import Leaderboard, User
from apps.desktop.data.scraper import fetch_global_leaderboard, fetch_monthly_leaderboard
from apps.desktop.ui.api.client import NookAPI

class LocalNookAPI(NookAPI):
    last_call: dict[Leaderboard, float] = dict()

    def get_leaderboard(self, leaderboard: Leaderboard) -> list[User]:
        now = time.time()

        if (
            leaderboard in self.last_call and
            (now - self.last_call[leaderboard]) < config.API_RATE_LIMIT_THRESHOLD_MS / 1000
        ):
            raise InterruptedError(config.ERR_API_RATE_LIMITED)
        self.last_call[leaderboard] = now

        if leaderboard == Leaderboard.GLOBAL:
            return fetch_global_leaderboard()
        elif leaderboard == Leaderboard.MONTHLY:
            return fetch_monthly_leaderboard()
        else:
            raise ValueError(
                config.ERR_API_INVALID_TYPE.format(type=leaderboard)
            )

    def refresh(self, leaderboard: Leaderboard) -> float:
        pass  # no-op since there's no cache to refresh
