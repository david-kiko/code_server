/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  env: {
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || '/api',
    NEXT_PUBLIC_WS_HOST: process.env.NEXT_PUBLIC_WS_HOST || 'localhost:8080',
    NEXT_PUBLIC_WS_PATH: process.env.NEXT_PUBLIC_WS_PATH || '/ws',
    NEXT_PUBLIC_APP_NAME: process.env.NEXT_PUBLIC_APP_NAME || '容器编排管理平台',
    NEXT_PUBLIC_VERSION: process.env.NEXT_PUBLIC_VERSION || '1.0.0',
  },
    webpack: (config, { isServer }) => {
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
        net: false,
        tls: false,
      };
    }
    return config;
  },
};

module.exports = nextConfig;