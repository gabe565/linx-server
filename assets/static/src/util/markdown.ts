import { Marked } from "marked";
import markedAlert from "marked-alert";

export const marked = new Marked().use(markedAlert());
