import { initAPI } from "./api.js";
import { initBin } from "./bin";
import "./main.css";
import "./paste";
import { initPaste } from "./paste";
import { initUpload } from "./upload";

if (document.querySelector("#dropzone")) {
  initUpload();
}

if (document.querySelector(".editor")) {
  initPaste();
}

if (document.querySelector("#normal-content")) {
  initBin();
}

if (document.querySelector("#main.api")) {
  initAPI();
}
