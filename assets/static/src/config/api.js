const base = window.config.site_path.replace(/\/+$/, "");

export const ApiPath = (path = "") => {
  const u = new URL(path, window.location.origin);
  if (path.match(/^\//)) {
    u.pathname = base + u.pathname;
  }
  return u.toString();
};
