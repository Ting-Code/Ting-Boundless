/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_DEV_LOGIN?: string;
  readonly VITE_SIGN_IN_PATH?: string;
  readonly VITE_GATEWAY_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
