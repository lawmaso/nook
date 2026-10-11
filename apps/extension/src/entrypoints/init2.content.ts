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
    
            const table = document.querySelector(".tinytext") as HTMLTableElement;
            // const recordCell = table.rows[2].cells[1];
            
            // const recordStr = recordCell.textContent;
            // const recordTime = Number(recordStr.split("\u00A0")[0]);
            // await chrome.storage.local.set({ record: recordTime });

            if (button) {
                button.click();
            } else {
                console.error("[nook] No start button found");
            }
        }, state[INIT2].min * 1000);
    }
});
