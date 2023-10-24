export default {
  "*.{js,ts,jsx,tsx}": [
    "prettier --write --ignore-unknown",
    "eslint --fix",
    () => "tsc -p tsconfig.json --noEmit",
  ],
};
