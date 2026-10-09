import { bootstrap } from "./api";
import { connectEvents } from "./events";
import { applySettings } from "./settings";
import { initTheme } from "./theme";

export async function startApp(): Promise<void> {
  initTheme();
  try {
    const hello = await bootstrap();
    applySettings(hello.settings as Record<string, unknown>);
    connectEvents();
  } catch {
    // SessionEnded or offline; shell shows relaunch screen when 401.
  }
}
