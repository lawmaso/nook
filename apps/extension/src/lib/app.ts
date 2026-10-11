export const SITE_URL     = "starbattle.puzzlebaron.com";
export const SITE_PATTERN = `*://${SITE_URL}/`;

export const INIT     = "init"     as const;
export const INIT2    = "init2"    as const;
export const GAMEOVER = "gameover" as const;

export const PAGE_NAMES = [INIT, INIT2, GAMEOVER] as const;
export type PageName = typeof PAGE_NAMES[number];

export const createUrl = (page: PageName) => {
    return `${SITE_PATTERN}${page}.php*`
}

export type PageConfig = {
    on?: boolean,
    min: number;
    max: number;
};

// The min and max wait times on each .php page
export const PAGE_BOUNDS: Record<PageName, PageConfig> = {
    init:     { min:  0, max: 10 },
    init2:    { min:  0, max: 10 },
    gameover: { min:  0, max: 10 }
};


// The chosen min and max wait times on each .php page
export type State = {
    on:       boolean,
    nooks:    number,
    init:     PageConfig,
    init2:    PageConfig
    gameover: PageConfig
};

export const DEFAULT_STATE: State = {
    on: true,
    nooks: 3,
    init:     { on: true, min:  0, max: 2   },
    init2:    { on: true, min:  0, max: 2   },
    gameover: { on: true, min:  0, max: 2   }
};

export type Message =
    | { type: "GET_STATE" }
    | { type: "SET_ON", on: boolean }
    | { type: "UPDATE_NOOKS", nooks: number }
    | { type: "UPDATE_PAGE_CONFIG", page: PageName, config: Partial<PageConfig> }
    | { type: "RESET_STATE" }
    | { type: "SOLVE", n: number, nooks: number, shapes: Record<any, number[][]> };
