// Change tab key behavior to insert tab instead of change focus
export const handleTab = (e) => {
  if (e.keyCode === 9) {
    e.preventDefault();

    const el = e.target;
    const start = el.selectionStart;
    const end = el.selectionEnd;

    el.value = el.value.substring(0, start) + "\t" + el.value.substring(end);
    el.selectionStart++;
    el.selectionEnd++;
  }
};
