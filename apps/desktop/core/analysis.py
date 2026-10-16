from apps.desktop.core.models import User

def _get_lower_user(u1: User, u2: User) -> User:
    return u1 if u1.total_points < u2.total_points else u2

def point_gap(u1: User, u2: User) -> int:
    return abs(u2.total_points - u1.total_points)

def puzzle_gap(u1: User, u2: User) -> tuple[float]:
    lo = _get_lower_user(u1, u2)

    gap = point_gap(u1, u2)
    return (
        gap / 563,               # estimate based on max possible score
        gap / lo.average_points  # estimate using lower ranked user's avg
    )

def time_gap(u1: User, u2: User) -> float:
    lo = _get_lower_user(u1, u2)
    avg_puzzles = puzzle_gap(u1, u2)[1]

    return avg_puzzles * lo.average_time

