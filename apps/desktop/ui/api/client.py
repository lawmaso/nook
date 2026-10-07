from abc import ABC, abstractmethod
from core.models import Leaderboard, User

class NookAPI(ABC):
    @abstractmethod
    def get_leaderboard(self, leaderboard: Leaderboard) -> list[User]:
        pass
    
    @abstractmethod
    def refresh(self, leaderboard: Leaderboard) -> float:
        pass

def get_api() -> NookAPI:
    # NOTE: defer import to prevent circular import error
    from apps.desktop.ui.api.local import LocalNookAPI
    return LocalNookAPI()
