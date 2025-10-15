import { AdminForthDataTypes } from "adminforth";
import type { AdminForthResourceInput, AdminUser } from "adminforth";

async function allowedForSuperAdmin({
  adminUser,
}: {
  adminUser: AdminUser;
}): Promise<boolean> {
  return adminUser.dbUser.role === "superadmin";
}

export default {
  dataSource: "maindb",
  table: "sweeps",
  resourceId: "sweeps",
  label: "Sweeps",
  recordLabel: (r) => `${r.id}`,
  options: {
    allowedActions: {
      edit: false,
      delete: false,
    },
  },
  columns: [
    {
      name: "transaction_id",
      type: AdminForthDataTypes.STRING,
      primaryKey: true,
      showIn: { all: false },
    },
    {
      name: "sweeping_session_id",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
    {
      name: "error_message",
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
