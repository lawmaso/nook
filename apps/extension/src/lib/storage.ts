import { DEFAULT_STATE, State } from "./app";

const STORAGE_KEY = "APP";

export async function getState(): Promise<State> {
    const result = await browser.storage.sync.get(STORAGE_KEY);
    return (result[STORAGE_KEY] as State) ?? DEFAULT_STATE;
}

export async function setState(partial: Partial<State>): Promise<State> {
    const currState = await getState();
    const newState = {
        ...currState,
        ...partial
    };

    await browser.storage.sync.set({ [STORAGE_KEY]: newState });
    return newState;
}
