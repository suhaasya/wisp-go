/** Playwright API helpers (HTTP boundary). */

/** @param {import('@playwright/test').Page} page */
export async function fetchCsrf(page) {
  const res = await page.request.get("/api/v1/hello");
  if (!res.ok()) {
    throw new Error(`hello failed: ${res.status()}`);
  }
  const body = await res.json();
  return /** @type {string} */ (body.csrf);
}

/**
 * @param {import('@playwright/test').Page} page
 * @param {string} csrf
 * @param {string} method
 * @param {string} path
 * @param {unknown} [data]
 */
export async function apiJSON(page, csrf, method, path, data) {
  const res = await page.request.fetch(path, {
    method,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      "X-Wisp-CSRF": csrf,
    },
    data: data === undefined ? undefined : JSON.stringify(data),
  });
  return res;
}

/** @param {import('@playwright/test').Page} page @param {object} profile */
export async function createConnectionProfile(page, profile) {
  const csrf = await fetchCsrf(page);
  const res = await apiJSON(page, csrf, "POST", "/api/v1/connections", profile);
  if (!res.ok()) {
    const text = await res.text();
    throw new Error(`create profile: ${res.status()} ${text}`);
  }
  return /** @type {{ id: string }} */ (await res.json());
}

/** @param {import('@playwright/test').Page} page @param {string} id */
export async function deleteConnectionProfile(page, id) {
  const csrf = await fetchCsrf(page);
  const res = await apiJSON(page, csrf, "DELETE", `/api/v1/connections/${encodeURIComponent(id)}`);
  if (res.status() !== 204 && !res.ok()) {
    throw new Error(`delete profile: ${res.status()}`);
  }
}
