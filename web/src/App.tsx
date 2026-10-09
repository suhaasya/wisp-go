import { useEffect, useState } from "preact/hooks";
import type { ComponentType } from "preact";
import { setThemePreference, themePreference, type ThemePreference } from "./theme";

const themeOptions: { value: ThemePreference; label: string }[] = [
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
  { value: "system", label: "System" },
];

function isGalleryRoute(): boolean {
  return location.hash === "#/gallery" || location.hash.startsWith("#/gallery?");
}

function Shell() {
  return (
    <div class="min-h-screen bg-canvas text-ink">
      <header class="flex flex-wrap items-center justify-between gap-3 border-b border-line bg-panel px-4 py-3">
        <h1 class="text-15 font-semibold text-ink">Wisp</h1>
        <div
          class="inline-flex rounded-control border border-line bg-panel p-1"
          role="group"
          aria-label="Theme"
        >
          {themeOptions.map((opt) => (
            <button
              key={opt.value}
              type="button"
              class={`rounded-control px-3 py-1 text-12-5 font-medium ${
                themePreference.value === opt.value
                  ? "bg-ink text-window"
                  : "text-ink-2 hover:bg-raise"
              }`}
              aria-pressed={themePreference.value === opt.value}
              onClick={() => setThemePreference(opt.value)}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </header>
      <main class="mx-auto max-w-3xl p-6">
        <section class="rounded-card border border-line bg-panel p-4">
          <p class="text-ink-2">
            Design tokens: <span class="text-ink">panel</span>,{" "}
            <span class="text-accent">accent</span>,{" "}
            <span class="text-env-prod">env-prod</span>.
          </p>
          {import.meta.env.DEV ? (
            <p class="mt-3 text-12-5 text-ink-3">
              Dev: open{" "}
              <a class="text-accent underline" href="#/gallery">
                #/gallery
              </a>{" "}
              for the component kit.
            </p>
          ) : null}
        </section>
      </main>
    </div>
  );
}

export function App() {
  const [Gallery, setGallery] = useState<ComponentType | null>(null);
  const [onGallery, setOnGallery] = useState(isGalleryRoute());

  useEffect(() => {
    const sync = () => setOnGallery(isGalleryRoute());
    window.addEventListener("hashchange", sync);
    sync();
    return () => window.removeEventListener("hashchange", sync);
  }, []);

  useEffect(() => {
    if (!import.meta.env.DEV || !onGallery) {
      setGallery(null);
      return;
    }
    let cancelled = false;
    import("./gallery/Gallery").then((mod) => {
      if (!cancelled) {
        setGallery(() => mod.Gallery);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [onGallery]);

  if (import.meta.env.DEV && onGallery && Gallery) {
    return <Gallery />;
  }

  return <Shell />;
}
