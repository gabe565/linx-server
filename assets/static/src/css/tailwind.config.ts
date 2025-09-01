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
          },
        },
      },
    },
  },
};
