import express from "express";
import AdminForth from "adminforth";
import usersResource from "./resources/adminuser.js";
import blockchainsResource from "./resources/blockchains.js";
import tokensResource from "./resources/tokens.js";
import transactionsResource from "./resources/transactions.js";
import sweepingSessionsResource from "./resources/sweeping_sessions.js";
import sweepsResource from "./resources/sweeps.js";
import walletsResource from "./resources/wallets.js";
import { fileURLToPath } from "url";
import path from "path";

const ADMIN_BASE_URL = "/admin";

export const admin = new AdminForth({
  baseUrl: ADMIN_BASE_URL,
  auth: {
    usersResourceId: "admin_users",
    usernameField: "email",
    passwordHashField: "password_hash",
    rememberMeDays: 30,
    loginBackgroundImage: "@@/assets/login-background.png",
  },
  customization: {
    brandName: "Fiatless",
    title: "Fiatless",
    favicon: "@@/assets/favicon.png",
    brandLogo: "@@/assets/logo.svg",
    datesFormat: "DD MMM",
    timeFormat: "HH:mm a",
    showBrandNameInSidebar: true,
    emptyFieldPlaceholder: "-",
    styles: {
      colors: {
        light: {
          primary: "#1a56db",
          sidebar: { main: "#f9fafb", text: "#213045" },
        },
        dark: {
          primary: "#82ACFF",
          sidebar: { main: "#1f2937", text: "#9ca3af" },
        },
      },
    },
  },
  dataSources: [
    {
      id: "maindb",
      url: `${process.env.DATABASE_URL}`,
    },
  ],
  resources: [
    usersResource,
    blockchainsResource,
    tokensResource,
    transactionsResource,
    sweepingSessionsResource,
    sweepsResource,
    walletsResource,
  ],
  menu: [
    {
      label: "Blockchains",
      icon: "hugeicons:blockchain-04",
      resourceId: "blockchains",
    },
    {
      label: "Tokens",
      icon: "token:btc",
      resourceId: "tokens",
    },
    {
      label: "Transactions",
      icon: "hugeicons:bitcoin-transaction",
      resourceId: "transactions",
    },
    {
      label: "Sweeping Sessions",
      icon: "tdesign:swap",
      resourceId: "sweeping_sessions",
    },
    {
      label: "Wallets",
      icon: "flowbite:wallet-solid",
      resourceId: "wallets",
    },
    { type: "heading", label: "SYSTEM" },
    {
      label: "Users",
      icon: "flowbite:user-solid",
      resourceId: "admin_users",
    },
  ],
});

if (fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) {
  const app = express();
  app.use(express.json());

  const port = 3500;

  admin
    .bundleNow({ hotReload: process.env.NODE_ENV === "development" })
    .then(() => {
      console.log("Bundling AdminForth SPA done.");
    });

  admin.express.serve(app);

  admin.discoverDatabases().then(async () => {
    if ((await admin.resource("admin_users").count()) === 0) {
      const email = process.env.ADMIN_EMAIL || "adminforth";
      const password = process.env.ADMIN_PASSWORD || "adminforth";
      await admin.resource("admin_users").create({
        email,
        password_hash: await AdminForth.Utils.generatePasswordHash(password),
        role: "superadmin",
      });
    }
  });

  admin.express.listen(port, () => {
    console.log(
      `\n⚡ AdminForth is available at http://localhost:${port}${ADMIN_BASE_URL}\n`
    );
  });
}
