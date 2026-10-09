import type { TableBrowseQuery } from "../state";
import { emptyBrowseQuery } from "./browse-query";

export type ForeignKeyMeta = {
  name: string;
  columns: string[];
  ref_schema: string;
  ref_table: string;
  ref_columns: string[];
};

export type ForeignKeyJump = {
  refSchema: string;
  refTable: string;
  linkLabel: string;
  browseQuery: TableBrowseQuery;
};

/** Maps local FK column values to a filtered browse on the referenced table. */
export function buildForeignKeyJump(
  fk: ForeignKeyMeta,
  localValues: Record<string, string | null | undefined>,
): ForeignKeyJump | null {
  const conditions: TableBrowseQuery["filter"]["conditions"] = [];
  for (let i = 0; i < fk.columns.length; i++) {
    const localCol = fk.columns[i]!;
    const refCol = fk.ref_columns[i]!;
    const raw = localValues[localCol];
    if (raw === null || raw === undefined || raw === "") {
      return null;
    }
    conditions.push({ column: refCol, op: "=", value: raw });
  }
  if (conditions.length === 0) {
    return null;
  }
  const last = conditions[conditions.length - 1]!.value;
  const browseQuery: TableBrowseQuery = {
    ...emptyBrowseQuery(),
    filter: { logic: "and", conditions },
  };
  return {
    refSchema: fk.ref_schema,
    refTable: fk.ref_table,
    linkLabel: `Open ${fk.ref_table} ${last}`,
    browseQuery,
  };
}
