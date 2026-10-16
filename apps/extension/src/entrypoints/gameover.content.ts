import { createUrl, GAMEOVER } from "@/lib/app";
import { getState } from "@/lib/storage";

export default defineContentScript({
    matches: [createUrl(GAMEOVER)],
    async main() {
        const state = await getState();
        if (!state.on || !state[GAMEOVER].on) {
            return;
        }

        setTimeout(() => {
            const button = document.querySelector(".button_green") as HTMLButtonElement;

            if (button) {
                button.click();
            } else {
                console.error("[nook] No create button found");
            }
        }, state[GAMEOVER].min * 1000);
    }
});
