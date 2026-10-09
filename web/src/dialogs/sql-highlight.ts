export type SqlTokenKind = "kw" | "str" | "num" | "plain";

export type SqlToken = { kind: SqlTokenKind; text: string };

const keywords = new Set([
  "UPDATE",
  "INSERT",
  "INTO",
  "DELETE",
  "FROM",
  "SET",
  "WHERE",
  "AND",
  "VALUES",
  "NULL",
  "IS",
]);

export function tokenizeSQL(sql: string): SqlToken[] {
  const out: SqlToken[] = [];
  let i = 0;
  const push = (kind: SqlTokenKind, text: string) => {
    if (text.length > 0) {
      out.push({ kind, text });
    }
  };
  while (i < sql.length) {
    const c = sql[i]!;
    if (c === "'") {
      let j = i + 1;
      while (j < sql.length) {
        if (sql[j] === "'" && sql[j + 1] === "'") {
          j += 2;
          continue;
        }
        if (sql[j] === "'") {
          break;
        }
        j++;
      }
      push("str", sql.slice(i, j + 1));
      i = j + 1;
      continue;
    }
    if (c >= "0" && c <= "9") {
      let j = i + 1;
      while (j < sql.length && ((sql[j]! >= "0" && sql[j]! <= "9") || sql[j] === ".")) {
        j++;
      }
      push("num", sql.slice(i, j));
      i = j;
      continue;
    }
    if ((c >= "A" && c <= "Z") || (c >= "a" && c <= "z") || c === "_") {
      let j = i + 1;
      while (j < sql.length) {
        const ch = sql[j]!;
        if (
          !(ch >= "A" && ch <= "Z") &&
          !(ch >= "a" && ch <= "z") &&
          !(ch >= "0" && ch <= "9") &&
          ch !== "_"
        ) {
          break;
        }
        j++;
      }
      const word = sql.slice(i, j);
      if (keywords.has(word.toUpperCase())) {
        push("kw", word);
      } else {
        push("plain", word);
      }
      i = j;
      continue;
    }
    push("plain", c);
    i++;
  }
  return out;
}

export function tokenClass(kind: SqlTokenKind): string {
  switch (kind) {
    case "kw":
      return "text-syn-kw font-medium";
    case "str":
      return "text-syn-str";
    case "num":
      return "text-syn-num";
    default:
      return "text-ink";
  }
}
