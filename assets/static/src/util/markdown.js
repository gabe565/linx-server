import DOMPurify from "dompurify";
import { marked } from "marked";

export default function parse(text) {
  return DOMPurify.sanitize(marked.parse(text));
}
