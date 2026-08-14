import colors from "tailwindcss/colors";

/** @type {import('tailwindcss').Config} */
export default {
  theme: {
    extend: {
      typography: {
        neutral: {
          css: {
            "--tw-prose-pre-code": colors.neutral[700],
            "--tw-prose-pre-bg": colors.neutral[50],
            "--tw-prose-bullets": colors.neutral[400],
          },
        },
        DEFAULT: {
          css: {
            "code::before": null,
            "code::after": null,
            code: {
              "border-radius": "0.25rem",
              "background-color": "var(--muted)",
              "padding-inline": "0.3rem",
              "padding-block": "0.2rem",
            },
            ".hljs": {
              display: "inline",
            },
            a: {
              "text-decoration": "none",
            },
            "a:hover": {
              "text-decoration": "underline",
            },
            "h1, h2": {
              "padding-bottom": "0.3em",
              "border-bottom": "1px solid var(--border)",
            },
            blockquote: {
              "font-style": "normal",
              "font-weight": "400",
              color: "var(--muted-foreground)",
              "border-inline-start-color": "var(--border)",
              quotes: "none",
            },
            "blockquote p:first-of-type::before": null,
            "blockquote p:last-of-type::after": null,
            hr: {
              height: "0.25em",
              border: "0",
              "background-color": "var(--border)",
            },
            kbd: {
              display: "inline-block",
              "font-size": "0.85em",
              "font-family": "inherit",
              "line-height": "1",
              "padding-inline": "0.4rem",
              "padding-block": "0.25rem",
              "background-color": "var(--muted)",
              border: "1px solid var(--border)",
              "border-radius": "0.375rem",
              "box-shadow": "inset 0 -1px 0 var(--border)",
            },
            'li:has(> input[type="checkbox"])': {
              "list-style-type": "none",
            },
            // Pull the box into the marker gutter so text lines up with plain items.
            'li > input[type="checkbox"]': {
              "margin-inline-start": "-1.4em",
              "margin-inline-end": "0.45em",
              "vertical-align": "middle",
            },
            // Table styling lives in main.css, applied to all prose tables.
            "thead, tbody tr": {
              "border-bottom": "0",
            },
          },
        },
      },
    },
  },
};
