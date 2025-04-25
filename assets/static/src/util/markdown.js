import DOMPurify from "dompurify";
import "github-markdown-css/github-markdown.css";
import { marked } from "marked";

export default function parse(text) {
  return DOMPurify.sanitize(marked.parse(text));
}
