import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  cacheComponents: true,
  partialPrefetching: true,
  cacheLife: {
    // The daemon scrapes monthly hourly and global every 2h, so a few minutes
    // of staleness is invisible while keeping Supabase traffic low.
    scrape: {
      stale: 60,
      revalidate: 300,
      expire: 3600,
    },
  },
  async redirects() {
    return [{ source: "/", destination: "/leaderboard/monthly", permanent: false }]
  },
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
