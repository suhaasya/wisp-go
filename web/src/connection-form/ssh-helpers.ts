export type SSHJumpForm = {
  host: string;
  port: number;
  user: string;
  auth: string;
  keyPath: string;
  configHost: string;
};

export type SSHFormState = {
  enabled: boolean;
  host: string;
  port: number;
  user: string;
  auth: string;
  keyPath: string;
  configHost: string;
  passphrase: string;
  passphraseDirty: boolean;
  password: string;
  passwordDirty: boolean;
  jumps: SSHJumpForm[];
};

export function defaultSSHForm(): SSHFormState {
  return {
    enabled: false,
    host: "",
    port: 22,
    user: "",
    auth: "key",
    keyPath: "",
    configHost: "",
    passphrase: "",
    passphraseDirty: false,
    password: "",
    passwordDirty: false,
    jumps: [],
  };
}

export const sshAuthMethods = [
  { value: "key", label: "Private key" },
  { value: "agent", label: "ssh-agent" },
  { value: "password", label: "Password" },
  { value: "config", label: "From ~/.ssh/config" },
];
