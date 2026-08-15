import nextConfig from "eslint-config-next/core-web-vitals";

const config = [
  ...nextConfig,
  {
    ignores: [".next/**", "coverage/**", "node_modules/**"],
  },
];

export default config;
