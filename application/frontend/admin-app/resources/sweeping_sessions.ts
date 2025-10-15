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
  table: "sweeping_sessions",
  resourceId: "sweeping_sessions",
  label: "Sweeping Sessions",
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
        create: false,
      },
    },
    {
      name: "wallet_meta_id",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "token_id",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "min_amount_threshold",
      type: AdminForthDataTypes.BOOLEAN,
      showIn: { all: false },
    },
    {
      name: "status",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "meta",
      type: AdminForthDataTypes.JSON,
      showIn: { all: false },
    },
    {
      name: "created_at",
      type: AdminForthDataTypes.DATETIME,
    },
  ],
} as AdminForthResourceInput;
