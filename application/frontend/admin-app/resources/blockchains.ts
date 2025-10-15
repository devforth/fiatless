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
  table: "blockchains",
  resourceId: "blockchains",
  label: "Blockchains",
  recordLabel: (r) => `🔗 ${r.name} (${r.network})`,
  options: {
    allowedActions: {
      edit: false,
      delete: false,
      create: false,
    },
  },
  columns: [
    {
      name: "id",
      primaryKey: true,
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
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
      showIn: {
        show: true,
        list: false,
        filter: false,
      },
    },
    {
      name: "network",
      type: AdminForthDataTypes.STRING,
      showIn: {
        show: true,
        list: true,
        filter: true,
      },
    },
    {
      name: "is_active",
      type: AdminForthDataTypes.BOOLEAN,
      showIn: { all: false },
    },
    {
      name: "logo_url",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "explorer_url",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "created_at",
      type: AdminForthDataTypes.DATETIME,
      showIn: { all: false },
    },
  ],
} as AdminForthResourceInput;
