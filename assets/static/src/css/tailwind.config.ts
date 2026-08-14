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
            // Class names come from marked-alert; variants only swap --alert-accent.
            ".markdown-alert": {
              "--alert-accent": "var(--border)",
              "margin-block": "1.25em",
              "padding-block": "0.5em",
              "padding-inline": "1em",
              "border-inline-start": "0.25em solid var(--alert-accent)",
            },
            ".markdown-alert > :first-child": { "margin-block-start": "0" },
            ".markdown-alert > :last-child": { "margin-block-end": "0" },
            ".markdown-alert-title": {
              display: "flex",
              "align-items": "center",
              "font-weight": "600",
              color: "var(--alert-accent)",
              "margin-block-end": "0.4em",
            },
            // Octicons ship without a fill, so they inherit the accent.
            ".markdown-alert-title .octicon": {
              fill: "currentColor",
              "margin-inline-end": "0.5em",
              "flex-shrink": "0",
            },
            ".markdown-alert-note": { "--alert-accent": "var(--alert-note)" },
            ".markdown-alert-tip": { "--alert-accent": "var(--alert-tip)" },
            ".markdown-alert-important": { "--alert-accent": "var(--alert-important)" },
            ".markdown-alert-warning": { "--alert-accent": "var(--alert-warning)" },
            ".markdown-alert-caution": { "--alert-accent": "var(--alert-caution)" },
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
