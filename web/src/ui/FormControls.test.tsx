import { fireEvent, render, screen } from "@testing-library/preact";
import { describe, expect, it, vi } from "vitest";
import { Checkbox } from "./Checkbox";
import { Input } from "./Input";
import { Select } from "./Select";

describe("Input", () => {
  it("associates label and emits value changes", () => {
    const onInput = vi.fn();
    render(<Input label="Host" value="localhost" onInput={onInput} />);
    const field = screen.getByLabelText("Host");
    fireEvent.input(field, { target: { value: "db.local" } });
    expect(onInput).toHaveBeenCalledWith("db.local");
  });

  it("supports password type", () => {
    render(<Input label="Password" type="password" />);
    expect(screen.getByLabelText("Password")).toHaveProperty("type", "password");
  });
});

describe("Select", () => {
  it("calls onChange with option value", () => {
    const onChange = vi.fn();
    render(
      <Select
        label="Engine"
        value="pg"
        options={[
          { value: "pg", label: "PostgreSQL" },
          { value: "my", label: "MySQL" },
        ]}
        onChange={onChange}
      />,
    );
    fireEvent.change(screen.getByLabelText("Engine"), { target: { value: "my" } });
    expect(onChange).toHaveBeenCalledWith("my");
  });
});

describe("Checkbox", () => {
  it("toggles via onChange", () => {
    const onChange = vi.fn();
    render(<Checkbox label="Use SSL" checked={false} onChange={onChange} />);
    fireEvent.click(screen.getByLabelText("Use SSL"));
    expect(onChange).toHaveBeenCalledWith(true);
  });
});
