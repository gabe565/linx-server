export const ApiPath = (path = "") => {
  const base = import.meta.env.VITE_API_ADDRESS || window.location.origin;
  return new URL(path, base).toString();
};
