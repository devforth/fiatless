import { AdminForthDataTypes } from "adminforth";
import type { AdminForthResourceInput, AdminUser } from "adminforth";
import { randomUUID } from "crypto";

async function allowedForSuperAdmin({
  adminUser,
}: {
  adminUser: AdminUser;
}): Promise<boolean> {
  return adminUser.dbUser.role === "superadmin";
}

export default {
  dataSource: "maindb",
  table: "tokens",
  resourceId: "tokens",
  label: "Tokens",
  recordLabel: (r) => `${r.name} (${r.symbol})`,
  options: {
    allowedActions: {
      edit: true,
      delete: false,
    },
  },
  columns: [
    {
      name: "id",
      primaryKey: true,
      type: AdminForthDataTypes.STRING,
      fillOnCreate: ({ initialRecord, adminUser }) => randomUUID(),
      showIn: {
        list: false,
        edit: false,
        create: false,
      },
    },
    {
      name: "token_id",
      type: AdminForthDataTypes.STRING,
      showIn: {
        list: false,
        edit: false,
        create: true,
      },
    },
    {
      name: "name",
      required: true,
      isUnique: true,
      type: AdminForthDataTypes.STRING,
    },
    {
      name: "symbol",
      type: AdminForthDataTypes.STRING,
    },
    {
      name: "is_active",
      type: AdminForthDataTypes.BOOLEAN,
      showIn: {
        show: true,
        list: true,
        filter: true,
        create: true,
        edit: true,
      },
      required: true,
    },
    {
      name: "type",
      type: AdminForthDataTypes.STRING,
      showIn: {
        show: false,
        list: false,
        filter: false,
      },
    },
    {
      name: "logo_url",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "yahoo_symbol",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "blockchain_id",
      type: AdminForthDataTypes.STRING,
      foreignResource: {
        resourceId: "blockchains",
      },
      showIn: {
        show: true,
        list: true,
        filter: true,
      },
    },
    {
      name: "created_at",
      type: AdminForthDataTypes.DATETIME,
      showIn: { all: false },
    },
  ],
} as AdminForthResourceInput;
