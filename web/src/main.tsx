import { render } from "preact";
import { App } from "./App";
import "./styles/app.css";
import { startApp } from "./main";

void startApp().then(() => {
  const root = document.getElementById("app");
  if (root) {
    render(<App />, root);
  }
});
