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
  table: "wallets",
  resourceId: "wallets",
  label: "Wallets",
  recordLabel: (r) => `${r.id}`,
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
      fillOnCreate: ({ initialRecord, adminUser }) => randomUUID(),
      showIn: {
        edit: false,
        list: false,
        create: false,
      },
    },
    {
      name: "address",
      type: AdminForthDataTypes.STRING,
      showIn: {
        show: true,
        list: true,
        filter: true,
      },
    },
    {
      name: "meta_id",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "index",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "derivation_path",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "created_at",
      type: AdminForthDataTypes.DATETIME,
    },
  ],
} as AdminForthResourceInput;
