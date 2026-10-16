import datetime as dt
import re
from decimal import Decimal, ROUND_DOWN
from zoneinfo import ZoneInfo

def uid(s: str) -> int:
    return int(s)

def rank(s: str) -> int:
    return int(s.split(".")[0])

def date(s: str, format: str) -> dt.datetime:
    return dt.datetime.strptime(s.upper(), format) \
        .replace(tzinfo=ZoneInfo("America/New_York"))

def solved(s: str) -> int:
    return int(s.replace(",", "").split(" ")[0])

def attempted(s: str) -> int:
    return int(s.replace(",", "").split(" ")[0])

def solved_attempted(s: str) -> tuple[int, int]:
    return (int(n) for n in s.split("/"))

def acceptance_rate(solved: int, attempted: int) -> float:
    # return round(float(solved / attempted) * 100, 2)
    n = Decimal((solved / attempted) * 100)
    n = n.quantize(Decimal("0.00"), rounding=ROUND_DOWN)
    return float(n)

_TIME_RE = re.compile(r"(?:(?:(\d+):)?(\d+):)?(\d+(?:\.\d+)?)", re.ASCII)
def avg_time_global(s: str) -> float:
    match = _TIME_RE.fullmatch(s.strip())
    if match is None:
        raise ValueError(f"invalid time {s!r}: expected [[h:]m:]s")

    hours, minutes, seconds = match.groups()
    return int(hours or 0) * 3600 + int(minutes or 0) * 60 + float(seconds)

def avg_time_monthly(s: str) -> float:
    return float(s.replace(",", "").split(" ")[0])

def avg_points(s: str) -> float:
    return float(s.replace(",", "").split(" ")[0])

def avg_points_global(total_points: int, solved: int) -> float:
    n = Decimal(total_points / solved)
    n = n.quantize(Decimal("0.00"), rounding=ROUND_DOWN)
    return float(n)

def total_points(s: str) -> int:
    return int(s.replace(",", ""))
