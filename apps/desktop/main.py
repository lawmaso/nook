from apps.desktop.ui.api.client import get_api
from apps.desktop.ui.app import Nook

if __name__ == "__main__":
    app = Nook(api=get_api())
    app.mainloop()
