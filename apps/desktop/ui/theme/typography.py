from dataclasses import dataclass, field


@dataclass
class FontSpec:
    family: str
    size:   int
    weight: str = "normal"

    def as_tuple(self) -> tuple:
        if self.weight != "normal":
            return (self.family, self.size, self.weight)
        return (self.family, self.size)
    
    def __call__(self) -> tuple:
        return self.as_tuple()


SEGOE = "Segoe UI"

@dataclass
class Typography:
    ui:         FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    12))
    heading:    FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    19, "bold"))
    header:     FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    19, "bold"))
    title:      FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    13, "bold"))
    small:      FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    12))
    label:      FontSpec = field(default_factory=lambda: FontSpec(SEGOE,     8, "bold"))
    button:     FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    10, "bold"))
    btn:        FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    10, "bold"))
    stat:       FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    12))
    stat_large: FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    22))
    star:       FontSpec = field(default_factory=lambda: FontSpec(SEGOE,    20, "bold"))
