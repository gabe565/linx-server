export const formatBytes = (bytes, { decimals = 0, hideUnit = false } = {}) => {
  if (!+bytes) return "0 B";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  let val = parseFloat((bytes / Math.pow(k, i)).toFixed(dm));
  if (!hideUnit) {
    val += ` ${sizes[i]}`;
  }
  return val;
};

export const formatBitsPerSecond = (bytesPerSecond, { decimals = 1 } = {}) => {
  if (!+bytesPerSecond) return "0 bps";

  const k = 1000;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["bps", "Kbps", "Mbps", "Gbps", "Tbps"];

  const bitsPerSecond = bytesPerSecond * 8;
  const i = Math.floor(Math.log(bitsPerSecond) / Math.log(k));

  return `${parseFloat((bitsPerSecond / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
};
