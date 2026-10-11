import { createUrl, INIT } from "@/lib/app";
import { getState } from "@/lib/storage";

export default defineContentScript({
    matches: [createUrl(INIT)],
    async main() {
        const state = await getState();
        if (!state.on || !state[INIT].on) {
            return;
        }

        const selectDiv = document.querySelector(".select-style");
        if (!selectDiv) {
            console.error("[nook] No select div found");
            return;
        }

        const select = selectDiv.firstChild as HTMLSelectElement;
        const options = select.options;

        if (!options) {
            console.error("[nook] No options found");
            return;
        }

        let found = false;
        for (const opt of options) {
            const optText = opt.text.toLowerCase();

            // const boardMatch = optText.includes("14x14");
            // const nookMatch = optText.includes("3 stars");
            const nookMatch = optText.includes(`${state.nooks} star${state.nooks > 1 ? "s" : ""}`);


            // if (boardMatch && nookMatch) {
            if (nookMatch) {
                select.value = opt.value;
                found = true;
                break;
            }
        }

        if (!found) {
            console.error("[nook] No option match found");
            return;
        }

        const button = document.querySelector(".button_green") as HTMLButtonElement;
        if (!button) {
            console.error("[nook] No creation button found");
        }

        setTimeout(() => button.click(), state[INIT].min * 1000);
    }
});
