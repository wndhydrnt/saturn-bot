function formatDateTime(d) {
  d.querySelectorAll(".datetime").forEach((e) => {
    let d = new Date(Number(e.textContent) * 1000);
    e.textContent = Intl.DateTimeFormat(navigator.language, {
      dateStyle: "short",
      timeStyle: "long",
    }).format(d);
  });
}
