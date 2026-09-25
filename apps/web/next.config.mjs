/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  webpack: (config, { dev }) => {
    if (dev) {
      // Disable Webpack persistent filesystem cache in development to prevent
      // ENOENT vendor-chunks desynchronization across pnpm symlinks and restarts.
      config.cache = false;
    }
    return config;
  },
};

export default nextConfig;
