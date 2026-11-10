export const USER_COLORS = [
    "#5DAFBF",  // ocean blue
    "#57A982",  // leaf green
    "#86C5D2",  // sky blue
    "#88B99A",  // sage green
    "#397F91",  // deep ocean
    "#B2CE9C",  // grass green
    "#76AFC0",  // coastal blue
    "#6E9E7B",  // forest green
    "#C5B99D",  // island sand
];

export const userColor = (i: number) =>
    USER_COLORS[i % USER_COLORS.length] ?? "#5DAFBF";
