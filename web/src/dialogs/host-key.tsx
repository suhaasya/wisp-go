import { Button } from "../ui/Button";
import { Modal } from "../ui/Modal";

export type HostKeyPrompt = {
  host: string;
  port: number;
  key_type: string;
  fingerprint: string;
};

export function HostKeyDialog({
  open,
  prompt,
  onTrust,
  onCancel,
}: {
  open: boolean;
  prompt: HostKeyPrompt | null;
  onTrust: () => void;
  onCancel: () => void;
}) {
  if (!prompt) {
    return null;
  }
  return (
    <Modal open={open} title="Trust SSH host key?" onClose={onCancel}>
      <p class="mb-3 text-13 text-ink-2">
        The host key for {prompt.host} is not in your known_hosts file.
      </p>
      <p class="mb-1 text-12-5 font-medium text-ink-2">Fingerprint</p>
      <p class="mb-4 font-mono text-12-5 text-ink">{prompt.fingerprint}</p>
      <div class="flex justify-end gap-2">
        <Button label="Cancel" onClick={onCancel} />
        <Button label="Trust and connect" variant="primary" onClick={onTrust} />
      </div>
    </Modal>
  );
}
