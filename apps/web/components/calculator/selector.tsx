"use client"

import {
    Combobox,
    ComboboxContent,
    ComboboxEmpty,
    ComboboxInput,
    ComboboxItem,
    ComboboxList,
} from "@/components/ui/combobox"
import { LeaderboardEntry } from "@/lib/models"

type SelectorProps = {
    entries: LeaderboardEntry[]
    value: LeaderboardEntry | null
    onValueChange: (entry: LeaderboardEntry | null) => void
    placeholder?: string
}

export function Selector({
    entries,
    value,
    onValueChange,
    placeholder = "Select a user"
}: SelectorProps) {
    return (
        <Combobox
            items={entries}
            value={value}
            onValueChange={onValueChange}
            itemToStringLabel={(entry) => `${entry.rank}. ${entry.user.username}`}
            itemToStringValue={(entry) => String(entry.uid)}
            isItemEqualToValue={(a, b) => a.uid === b.uid}
        >
            <ComboboxInput placeholder={placeholder}/>
            <ComboboxContent>
                <ComboboxEmpty>No users found</ComboboxEmpty>
                <ComboboxList>
                    {(entry: LeaderboardEntry) => (
                        <ComboboxItem key={entry.uid} value={entry}>
                            {`${entry.rank}. ${entry.user.username}`}
                        </ComboboxItem>
                    )}
                </ComboboxList>
            </ComboboxContent>
        </Combobox>
    )
}
