/// <reference types="vite/client" />

// Allow importing of CSS files in TypeScript modules
declare module '*.css';

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

