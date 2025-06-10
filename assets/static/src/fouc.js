(() => {
  const darkMode = localStorage["vueuse-color-scheme"] || "auto";
  const force =
    darkMode === "dark" ||
    (darkMode === "auto" && window.matchMedia("(prefers-color-scheme: dark)").matches);
  document.documentElement.classList.toggle("dark", force);
})();
