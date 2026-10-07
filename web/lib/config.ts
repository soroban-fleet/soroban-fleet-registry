export interface AppConfig {
  apiBaseUrl: string;
  network: string;
  explorerBaseUrl: string;
}

export const config: AppConfig = {
  apiBaseUrl: process.env.NEXT_PUBLIC_SFR_API_URL || process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  network: process.env.NEXT_PUBLIC_NETWORK || 'testnet',
  explorerBaseUrl: process.env.NEXT_PUBLIC_EXPLORER_BASE_URL || 'https://stellar.expert/explorer/testnet',
};
