import { bootstrap } from "./api";
import { connectEvents } from "./events";
import { initTheme } from "./theme";

export async function startApp(): Promise<void> {
  initTheme();
  try {
    await bootstrap();
    connectEvents();
  } catch {
    // SessionEnded or offline; shell shows relaunch screen when 401.
  }
}
