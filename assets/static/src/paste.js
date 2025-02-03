import { handleTab } from "./util";

export const initPaste = () => {
  document.querySelector(".editor").addEventListener("keydown", handleTab);
};
