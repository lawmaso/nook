def format_large(n: int) -> str:
    return f"{n:,}"

def point_gap(gap: int) -> str:
    return format_large(gap)

def puzzle_gap(range_: list[float]) -> str:
    perfect, expected = range_
    return f"{perfect:,.0f}-{expected:,.0f}" if \
        (perfect < 1e6 or expected < 1e6) else "∞"

def time_gap(seconds: float) -> str:
    seconds = int(seconds)

    y = seconds // (3600 * 24 * 365)    # years
    d = (seconds // (3600 * 24)) % 365  # days 
    h = (seconds // 3600) % 24          # hours
    m = (seconds % 3600) // 60          # minutes
    s = seconds % 60                    # seconds

    time_string = ""
    if y > 0: time_string += f"{y}y "
    if d > 0: time_string += f"{d}d "
    if h > 0: time_string += f"{h}h "
    if m > 0: time_string += f"{m}m "
    time_string += f"{str(s).zfill(1)}s"

    return time_string

def seconds(sec: float) -> str:
    return f"{format_large(sec)}s"

def percent(pct: float) -> str:
    return f"{format_large(pct)}%"
