import { WISP_FEEDBACK_URL, WISP_USER_DOCS_URL } from "../../feedback";
import { serverVersion } from "../../state";

export function AboutPane() {
  return (
    <div class="flex flex-col gap-4 text-13 text-ink">
      <p class="text-ink-2">
        Wisp is a local database client for PostgreSQL, MySQL and MariaDB. Data stays on your machine; the UI runs in
        your browser and talks to a private loopback server.
      </p>
      <dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-12-5">
        <dt class="text-ink-2">Version</dt>
        <dd class="font-mono text-ink">{serverVersion.value}</dd>
      </dl>
      <div class="flex flex-wrap gap-2">
        <a
          href={WISP_FEEDBACK_URL}
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center rounded-control border border-line bg-raise px-3 py-1.5 font-medium text-ink hover:bg-panel"
        >
          Send feedback
        </a>
        <a
          href={WISP_USER_DOCS_URL}
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center rounded-control border border-line bg-transparent px-3 py-1.5 text-ink-2 hover:text-ink"
        >
          User guide
        </a>
      </div>
      <p class="text-12-5 text-ink-2">
        During beta, please report crashes, connection failures and UI bugs via Send feedback. Include your OS, browser
        and database engine version.
      </p>
    </div>
  );
}
