import { Message, PageConfig, PageName, State } from "@/lib/app";
import { useEffect, useState } from "react";

export function useNookState() {
    const [state, setState] = useState<State | null>(null);

    useEffect(() => {
        browser.runtime.sendMessage({ type: "GET_STATE" } satisfies Message).then(setState);
    }, []);

    async function send(message: Message) {
        const newState: State = await browser.runtime.sendMessage(message);
        setState(newState);
        return newState;
    };

    return {
        state,
        setOn: (on: boolean) => send({ type: "SET_ON", on }),
        updateNooks: (nooks: number) => send({ type: "UPDATE_NOOKS", nooks }),
        updatePageConfig: (page: PageName, config: Partial<PageConfig>) =>
            send({ type: "UPDATE_PAGE_CONFIG", page, config }),
        resetBounds: () => send({ type: "RESET_STATE" }),
        setLocalPageConfig: (page: PageName, config: Partial<PageConfig>) =>
            setState((prev) => (prev
                ? { ...prev, [page]: { ...prev[page], ...config }}
                : prev
            ))
    };
}
