import { render } from "preact";
import { App } from "./App";
import "./styles/app.css";
import { initTheme } from "./theme";

initTheme();

const root = document.getElementById("app");
if (root) {
  render(<App />, root);
}
