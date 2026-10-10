import path from "node:path";
import { fileURLToPath } from "node:url";

const e2eDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.join(e2eDir, "..");
export const ciDir = path.join(repoRoot, "ci");
export const sshClientKey = path.join(ciDir, "ssh", "client", "id_ed25519");
export const tlsCA = path.join(ciDir, "certs", "ca.pem");
