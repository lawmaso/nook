import datetime as dt
from dataclasses import dataclass
from enum import Enum

class Leaderboard(Enum):
    MONTHLY = "monthly"
    GLOBAL  = "global"

@dataclass
class User:
    uid:             int
    rank:            int
    username:        str
    last_active:     dt.datetime
    solved:          int
    attempted:       int
    acceptance_rate: float
    average_time:    float
    average_points:  float
    total_points:    int
    
    def __str__(self) -> str:
        rank = f"{self.rank:>3}"
        username = f"{'@{}'.format(self.username):<33}"
        total_points = f"{'{:,} {}'.format(self.total_points, 'pts'):<15}"

        return f"{rank} {username} {total_points}"

    @property
    def row_values(self) -> tuple[str]:
        u = self
        return (
            f"#{u.rank}",
            f"{u.username}",
            f"{u.total_points:,}",
            f"{u.solved}/{u.attempted}",
            f"{u.average_time:.2f}s",
            f"{u.acceptance_rate:.2f}%"
        )


if __name__ == "__main__":
    airmaso = User(
        uid=0, rank=1, username="airmaso", last_active=dt.datetime.now(),
        solved=1919, attempted=1919, acceptance_rate=100.00,
        average_time=190.00, average_points=555.00, total_points=1_800_000
    )
    print(airmaso)
