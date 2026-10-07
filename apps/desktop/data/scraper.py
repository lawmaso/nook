import logging
import threading
import requests
from bs4 import BeautifulSoup
from apps.desktop.core.models import User, Leaderboard
from apps.desktop.data import parsers as p

logger = logging.getLogger(__name__)
GLOBAL_URL  = "https://starbattle.puzzlebaron.com/halloffame.php"
MONTHLY_URL = "https://starbattle.puzzlebaron.com/contest1.php"
PLAYER_URL  = "https://starbattle.puzzlebaron.com/profile.php?u={uid}"


def _get_leaderboard_data(type: Leaderboard) -> list[list[str]] | None:
    try:
        url = GLOBAL_URL if type == Leaderboard.GLOBAL else MONTHLY_URL

        response = requests.get(url, timeout=10)
        response.raise_for_status()

        soup = BeautifulSoup(response.text, "html.parser")
        table = soup.find("table", class_="winners")
        rows = table.find_all("tr")[1:]  # [1:] to skip header

        data = []
        for row in rows:
            cols = [e.text.strip() for e in row.find_all("td")]
            anchor = row.find_next("a")
            uid = anchor.attrs["href"].split("?u=")[1]

            data.append([uid] + cols)

        return data
    except Exception as e:
        logger.error("Failed to fetch leaderboard data: %s", e)
        return None


def _get_additional_data(uid: int) -> list[list[str]] | None:
    try:
        response = requests.get(PLAYER_URL.format(uid=uid), timeout=10)
        response.raise_for_status()

        soup = BeautifulSoup(response.text, "html.parser")

        scorecard_tables = soup.find_all("table", class_="scorecard_table")
        scorecard_table = scorecard_tables[5]
        lifetime_table = soup.find("div", id="tabs-1")

        scorecard_rows = scorecard_table.find_all("tr")
        avg_time = scorecard_rows[3].find_all("td")[1].text.strip()
        lifetime_rows = lifetime_table.find_all("tr")[1:]  # [1:] skip header

        data = []
        for row in lifetime_rows:
            cols = [e.text.strip() for e in row.find_all("td")][1]
            data.append(cols)
        
        data.append(avg_time)
        return data
    except Exception as e:
        logger.error("Failed to fetch additional data for uid=%s: %s",
                     uid, e)
        return None


def fetch_global_leaderboard() -> list[User | None]:
    # global users is generally fixed, so we can use a fixed-size list
    # this is not the case for monthly users due to leaderboard refreshes
    global_users: list[User | None] = [None] * 101
    failed = threading.Event()  # shared error flag

    data = _get_leaderboard_data(Leaderboard.GLOBAL)
    if data is None:
        return []

    threads = []
    def process_user(row: list[str]):
        if failed.is_set():
            return

        uid, rank, username, last_active, total_points = row
        additional_data = _get_additional_data(int(uid))

        if additional_data is None:
            failed.set()
            return

        member_since, _, attempted, \
            solved, acceptance_rate, avg_points, avg_time = additional_data

        uid                 = p.uid(uid)
        rank                = p.rank(rank)
        last_active         = p.date(last_active, "%B %d, %Y")
        member_since        = p.date(member_since, "%B %d, %Y")
        solved              = p.solved(solved)
        attempted           = p.attempted(attempted)
        acceptance_rate     = p.acceptance_rate(solved, attempted)
        avg_time            = p.avg_time_global(avg_time)
        avg_points          = p.avg_points(avg_points)
        total_points        = p.total_points(total_points)

        global_users[rank] = User(
            uid=uid, rank=rank, username=username,
            last_active=last_active, solved=solved, attempted=attempted,
            acceptance_rate=acceptance_rate, average_time=avg_time,
            average_points=avg_points, total_points=total_points
        )

    threads = [threading.Thread(target=process_user, args=(row,)) for row in data]
    for t in threads: t.start()
    for t in threads: t.join()

    if failed.is_set():
        logger.error("Failed to fetch all global users")
        return []

    return global_users


def fetch_monthly_leaderboard() -> list[User | None]:
    # monthly_users: list[User | None] = [None] * 101  # dangerous on monthly refreshes
    monthly_users: list[User | None] = [None]  # initial sentinel user
    data = _get_leaderboard_data(Leaderboard.MONTHLY)

    if data is None:
        logger.error("Failed to fetch all monthly users")
        return []

    for row in data:
        uid, rank, username, last_active, solved_attempted, \
            acceptance_rate, avg_time, avg_points, total_points = row

        uid                 = p.uid(uid)
        rank                = p.rank(rank)
        last_active         = p.date(last_active, "%B %d, %Y, %I:%M %p")
        solved, attempted   = p.solved_attempted(solved_attempted)
        acceptance_rate     = p.acceptance_rate(solved, attempted)
        avg_time            = p.avg_time_monthly(avg_time)
        avg_points          = p.avg_points(avg_points)
        total_points        = p.total_points(total_points)

        monthly_users.append(User(
            uid=uid, rank=rank, username=username,
            last_active=last_active, solved=solved, attempted=attempted,
            acceptance_rate=acceptance_rate, average_time=avg_time,
            average_points=avg_points, total_points=total_points
        ))

    return monthly_users


if __name__ == "__main__":
    import time

    logging.basicConfig(
        filename="data/scraper.log",
        format="(%(asctime)s %(levelname)s %(message)s)",
        level=logging.INFO
    )

    for leaderboard, fetch_users in [
        (Leaderboard.MONTHLY, fetch_monthly_leaderboard),
        (Leaderboard.GLOBAL, fetch_global_leaderboard)
    ]:
        start = time.time()
        logger.info("Fetching %s users", leaderboard.value)
        users = fetch_users()
        logger.info("Fetched %s %s users in %.2fs",
                    len(users) - 1, leaderboard.value,
                    time.time() - start)
        logger.info("%s users:\n%s",
                    leaderboard.value.capitalize(),
                    "\n".join(str(u) for u in users[1:]))
