import { createUrl, INIT } from "@/lib/app";
import { getState } from "@/lib/storage";

export default defineContentScript({
    matches: [createUrl(INIT)],
    async main() {
        const state = await getState();
        if (!state.on || !state[INIT].on) {
            return;
        }

        function selectNooks(stars: number): boolean {
            const STAR_COUNT_SELECTOR = "span > i";

            const classic = document.querySelector<HTMLFieldSetElement>("fieldset.sbp-sizes");
            if (!classic) {
                console.error("[nook] No fieldset found");
                return false;
            }

            const chips = classic.querySelectorAll<HTMLLabelElement>("label.sbp-chip");
            if (!chips) {
                console.error("[nook] No chips found");
                return false;
            }

            const match = Array.from(chips).find((chip) => {
                const text = chip.querySelector(STAR_COUNT_SELECTOR)?.textContent;
                return text?.length == stars;
            });

            const input = match?.querySelector<HTMLInputElement>("input[type='radio']");
            if (!input) {
                console.error(`[nook] No size chip found for ${stars} star(s)`);
                return false;
            }

            input.checked = true;
            return true;
        }

        const found = selectNooks(state.nooks);
        if (!found) {
            console.error("[nook] No nook selection found");
            return;
        }

        const button = document.getElementById("sbp-go") as HTMLButtonElement;
        if (!button) {
            console.error("[nook] No creation button found");
            return;
        }

        setTimeout(() => button.click(), state[INIT].min * 1000);
    }
});
