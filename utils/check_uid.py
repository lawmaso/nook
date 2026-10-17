import requests
from bs4 import BeautifulSoup
from apps.desktop.data.scraper import PLAYER_URL

if __name__ == "__main__":
    lower_bound = 10_866
    upper_bound = 10_888

    for uid in range(lower_bound, upper_bound + 1):
        response = requests.get(PLAYER_URL.format(uid=uid))
        soup = BeautifulSoup(response.text, "html.parser")

        container = soup.find("div", id="container_left")
        style = container.find_all("style")

        if len(style) != 0:
            header = container.find("h1", class_="header_font").text.strip()
            username = header.split("'")[0]
            print(f"uid={uid} taken by @{username}")
        else:
            print(f"uid={uid} is available")
            break
