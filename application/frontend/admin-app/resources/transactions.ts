import { AdminForthDataTypes, Filters } from "adminforth";
import type { AdminForthResourceInput, AdminUser } from "adminforth";
import { randomUUID } from "crypto";
import { admin } from "../index.js";

async function allowedForSuperAdmin({
  adminUser,
}: {
  adminUser: AdminUser;
}): Promise<boolean> {
  return adminUser.dbUser.role === "superadmin";
}

export default {
  dataSource: "maindb",
  table: "transactions",
  resourceId: "transactions",
  label: "Transactions",
  recordLabel: (r) => `💰 ${r.id}`,
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
        show: true,
        list: false,
        filter: true,
        edit: false,
        create: false,
      },
    },
    {
      name: "tx_id",
      type: AdminForthDataTypes.STRING,
      showIn: {
        edit: false,
        create: false,
      },
      label: "Transaction ID",
      components: {
        list: '@/renderers/CompactUUID.vue'
      }
    },
    {
      name: "type",
      type: AdminForthDataTypes.STRING,
      showIn: {
        show: true,
        list: true,
        filter: true,
      },
    },
    {
      name: "to_address",
      type: AdminForthDataTypes.STRING,
      label: "Address",
    },
    {
      name: "amount",
      type: AdminForthDataTypes.STRING,
      showIn: { show: true, list: true},
    },
    {
      name: "fee",
      type: AdminForthDataTypes.DECIMAL,
      showIn: { show: true },
    },
    {
      name: "explorer_url",
      type: AdminForthDataTypes.STRING,
      showIn: { show: true },
      virtual: true,
      label: "Explorer",
    },
    {
      name: "token_id",
      type: AdminForthDataTypes.STRING,
      showIn: {
        show: false,
        list: false,
        filter: false,
      },
    },
    {
      name: "created_at",
      type: AdminForthDataTypes.STRING,
      showIn: { all: false },
    },
  ],
  hooks: {
    show: {
      afterDatasourceResponse: async ({ response, adminforth }) => {
        await Promise.all(response.map(async (record: any) => {
          const token = await adminforth.resource("tokens").get([Filters.EQ("id", record.token_id)]);
          if (token) {
            record.amount = parseFloat(record.amount).toString() + " " + token.symbol;
          }
        }));
        return { ok: true };
      },  
    },
    list: {
      afterDatasourceResponse: async ({ response, adminforth }) => {
        await Promise.all(response.map(async (record: any) => {
          const token = await adminforth.resource("tokens").get([Filters.EQ("id", record.token_id)]);
          if (token) {
            record.amount = parseFloat(record.amount).toString() + " " + token.symbol;
          }
        }));
        return { ok: true };
      },
    },
  },
} as AdminForthResourceInput;
