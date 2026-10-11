import { ModeToggle } from "@/components/mode-toggle";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { Kbd } from "@/components/ui/kbd";
import { PAGE_NAMES, PAGE_BOUNDS, INIT } from "@/lib/app";
import { Switch } from "@/components/ui/switch";
import {
    Field,
    FieldContent,
    FieldDescription,
    FieldGroup,
    FieldLabel,
    FieldTitle
} from "@/components/ui/field";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Map, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Rating } from "@/components/reui/rating";
import { useNookState } from "./hooks/useNookState";

export default function App() {
    const {
        state,
        setOn,
        updateNooks,
        updatePageConfig,
        resetBounds,
        setLocalPageConfig
    } = useNookState();

    function createBoundsString(min: number, max: number) {
        return `${min}-${max}s`
    }

    if (!state) {
        return;
    }

    return (
        <div className="flex flex-col mx-1 mb-1 py-3 px-2 gap-2.5">
            <div className="flex justify-between">
                <h1 className="flex items-center gap-1 text-lg font-semibold">
                    <Map className="text-primary" />
                    Nook
                </h1>
                <div className="flex gap-1">
                    <AlertDialog>
                        <AlertDialogTrigger
                            render={
                                <Button variant="outline" size="icon">
                                    <RotateCcw />
                                </Button>
                            }
                        />
                        <AlertDialogContent size="sm">
                            <AlertDialogHeader>
                                <AlertDialogTitle>Reset bounds?</AlertDialogTitle>
                                <AlertDialogDescription>
                                    This cannot be undone
                                </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <AlertDialogAction onClick={resetBounds}>Reset</AlertDialogAction>
                            </AlertDialogFooter>
                        </AlertDialogContent>
                    </AlertDialog>
                    <ModeToggle />
                </div>
            </div>

            <FieldGroup>
                <FieldLabel htmlFor="nook-toggle" className="cursor-pointer">
                    <Field orientation="horizontal">
                        <FieldContent>
                            <FieldTitle>On</FieldTitle>
                            <FieldDescription>Turn the app on or off</FieldDescription>
                        </FieldContent>
                        <Switch
                            id="nook-toggle"
                            checked={state.on}
                            onCheckedChange={() => setOn(!state.on)}
                            className="cursor-pointer"
                        />
                    </Field>
                </FieldLabel>
            </FieldGroup>

            {/* <pre>
                <code>{JSON.stringify(state, null, 2)}</code>
            </pre> */}

            <div className="flex flex-col gap-4 mx-0.5">
                {PAGE_NAMES.map((page) => (
                    <div key={page} className={`grid gap-3 ${(!state.on || !state[page].on) && "opacity-45"}`}>
                        <div className="flex items-center justify-between gap-2">
                            <div className="flex gap-1">
                                <Label htmlFor={`${page}-toggle`}>
                                    <Switch
                                        id={`${page}-toggle`}
                                        checked={state[page].on}
                                        onCheckedChange={() => updatePageConfig(page, { on: !state[page].on })}
                                        className="cursor-pointer"
                                    />
                                </Label>
                                <Kbd>{`${page}.php`}</Kbd>
                                {page === INIT && (
                                    <Rating
                                        rating={state.nooks}
                                        maxRating={3}
                                        editable
                                        onRatingChange={updateNooks}
                                        // size="sm"
                                    />
                                )}
                            </div>
                            <span className="text-xs text-muted-foreground">
                                {createBoundsString(state[page].min, state[page].max)}
                            </span>
                        </div>

                        <Slider
                            id={page}
                            value={[state[page].min, state[page].max]}
                            onValueChange={(value) => {
                                if (typeof value === "number") return;
                                setLocalPageConfig(page, { min: value[0], max: value[1] });
                            }}
                            onValueCommitted={(value) => {
                                if (typeof value === "number") return;
                                updatePageConfig(page, { min: value[0], max: value[1] });
                            }}
                            min={PAGE_BOUNDS[page].min}
                            max={PAGE_BOUNDS[page].max}
                            step={0.1}
                            className="cursor-pointer"
                        />
                    </div>
                ))}
            </div>
        </div>
    );
}
