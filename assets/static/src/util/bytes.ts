export const formatBytes = (bytes: number, decimals = 0) => {
  if (!+bytes) return "0 B";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${Number((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
};

export const formatBitsPerSecond = (bytes: number, decimals = 1) => {
  if (!+bytes) return "0 bps";

  const k = 1000;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["bps", "Kbps", "Mbps", "Gbps", "Tbps"];

  const bits = bytes * 8;
  const i = Math.floor(Math.log(bits) / Math.log(k));

  return `${Number((bits / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
};
