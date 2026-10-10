import net from "node:net";

/** @param {number} port */
export function portOpen(port) {
  return new Promise((resolve) => {
    const sock = net.connect({ host: "127.0.0.1", port }, () => {
      sock.end();
      resolve(true);
    });
    sock.on("error", () => resolve(false));
    sock.setTimeout(800, () => {
      sock.destroy();
      resolve(false);
    });
  });
}

/** @param {number[]} ports */
export async function matrixDockerReady(ports) {
  for (const p of ports) {
    if (!(await portOpen(p))) {
      return false;
    }
  }
  return true;
}
