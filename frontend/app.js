document.addEventListener("DOMContentLoaded", () => {
  const form = document.getElementById("order-form");
  const orderInput = document.getElementById("order-id-input");
  const errorEl = document.getElementById("form-error");
  const detailsEl = document.getElementById("details");

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    errorEl.textContent = "";
    detailsEl.textContent = "";

    const uid = orderInput.value.trim();
    if (!uid) {
      errorEl.textContent = "Введите Order UID";
      return;
    }

    try {
      const res = await fetch(`http://localhost:10000/order/${uid}`);
      if (!res.ok) {
        throw new Error(`Ошибка: ${res.status}`);
      }
      const data = await res.json();
      detailsEl.textContent = JSON.stringify(data, null, 2);
    } catch (err) {
      errorEl.textContent = err.message;
    }
  });
});
