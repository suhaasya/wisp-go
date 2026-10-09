import { render, screen } from "@testing-library/preact";
import { describe, expect, it } from "vitest";
import { ReviewSQLDialog } from "./review-sql";

describe("ReviewSQLDialog", () => {
  it("shows title and commit footer note", () => {
    render(
      <ReviewSQLDialog
        open={true}
        title="Review 2 changes to orders"
        subtitle="Local · PostgreSQL"
        sql="UPDATE t SET x = 1;"
        onClose={() => {}}
        onCommit={() => {}}
      />,
    );
    expect(screen.getByText("Review 2 changes to orders")).toBeTruthy();
    expect(screen.getByText("Nothing is written until you commit.")).toBeTruthy();
  });
});
