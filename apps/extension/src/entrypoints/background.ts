import { DEFAULT_STATE, Message } from "@/lib/app";
import { getState, setState } from "@/lib/storage";

export default defineBackground(() => {
    browser.runtime.onMessage.addListener(async (message: Message) => {
        switch (message.type) {
            case "GET_STATE": {
                return await getState();
            }

            case "SET_ON": {
                return await setState({ on: message.on });
            }

            case "UPDATE_NOOKS": {
                return await setState({ nooks: message.nooks });
            }

            case "UPDATE_PAGE_CONFIG": {
                const state = await getState();
                return await setState({
                    [message.page]: {
                        ...state[message.page],
                        ...message.config
                    }
                });
            }

            case "RESET_STATE": {
                return await setState(DEFAULT_STATE);
            }

            default: {
                console.error("[nook] Unknown message type:", (message as Message).type);
                return;
            }
        }
    });
});
