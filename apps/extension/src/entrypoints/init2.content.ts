import { createUrl, INIT2 } from "@/lib/app";
import { getState } from "@/lib/storage";

export default defineContentScript({
    matches: [createUrl(INIT2)],
    async main() {
        const state = await getState();
        if (!state.on || !state[INIT2].on) {
            return;
        }

        setTimeout(() => {
            const button = document.querySelector(".button_green") as HTMLButtonElement;
            if (!button) {
                console.error("[nook] No start button found");
            }

            button.click();
        }, state[INIT2].min * 1000);
    }
});
