import { useEffect } from "preact/hooks";
import { request } from "../../api";
import { formFromParseResult, type FormState } from "../../connection-form";
import { navigate } from "../../router";
import { pushToast } from "../../ui/Toast";

export const IMPORT_DRAFT_KEY = "wisp-import-draft";

export type ParseURLResponse = {
  profile: Record<string, unknown>;
  password?: string;
  warnings?: string[];
};

export function looksLikeConnectionURL(text: string): boolean {
  const s = text.trim();
	if (s === "") {
		return false;
	}
	const lower = s.toLowerCase();
	if (
		lower.startsWith("postgres://") ||
		lower.startsWith("postgresql://") ||
		lower.startsWith("mysql://") ||
		lower.startsWith("mariadb://")
	) {
		return true;
	}
	if (lower.includes("database_url=")) {
		return true;
	}
	if (lower.includes("host=") && (lower.includes("dbname=") || lower.includes("database="))) {
		return true;
	}
	return false;
}

export async function parseConnectionURL(text: string): Promise<{ form: FormState; warnings: string[] }> {
	const out = await request<ParseURLResponse>("POST", "/api/v1/connections/parse-url", { url: text });
	return { form: formFromParseResult(out.profile, out.password ?? ""), warnings: out.warnings ?? [] };
}

export function stashImportDraft(form: FormState): void {
	sessionStorage.setItem(IMPORT_DRAFT_KEY, JSON.stringify(form));
}

export function takeImportDraft(): FormState | null {
	const raw = sessionStorage.getItem(IMPORT_DRAFT_KEY);
	if (!raw) {
		return null;
	}
	sessionStorage.removeItem(IMPORT_DRAFT_KEY);
	try {
		return JSON.parse(raw) as FormState;
	} catch {
		return null;
	}
}

function isEditableTarget(target: EventTarget | null): boolean {
	if (!(target instanceof HTMLElement)) {
		return false;
	}
	const tag = target.tagName;
	if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") {
		return true;
	}
	return target.isContentEditable;
}

export function useConnectionPaste(enabled: boolean): void {
	useEffect(() => {
		if (!enabled) {
			return;
		}
		const onPaste = (e: ClipboardEvent) => {
			if (isEditableTarget(e.target)) {
				return;
			}
			const text = e.clipboardData?.getData("text/plain") ?? "";
			if (!looksLikeConnectionURL(text)) {
				pushToast("Not a connection URL", "error");
				return;
			}
			e.preventDefault();
			void parseConnectionURL(text)
				.then(({ form, warnings }) => {
					stashImportDraft(form);
					for (const w of warnings) {
						pushToast(w, "default");
					}
					navigate("/connections/new");
				})
				.catch(() => pushToast("Could not parse connection URL", "error"));
		};
		document.addEventListener("paste", onPaste);
		return () => document.removeEventListener("paste", onPaste);
	}, [enabled]);
}
