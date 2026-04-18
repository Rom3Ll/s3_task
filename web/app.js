(function () {
  const prefixInput = document.getElementById("prefix");
  const refreshBtn = document.getElementById("refresh");
  const tbody = document.getElementById("tbody");
  const statusEl = document.getElementById("status");
  const uploadBtn = document.getElementById("upload-local");
  const uploadStatusEl = document.getElementById("upload-status");
  const uploadLogEl = document.getElementById("upload-log");
  const copyFromInput = document.getElementById("copy-from");
  const copyToInput = document.getElementById("copy-to");
  const copyDeleteInput = document.getElementById("copy-delete");
  const copyBtn = document.getElementById("copy-prefix");
  const copyStatusEl = document.getElementById("copy-status");

  function formatBytes(n) {
    if (n === 0) return "0 B";
    const units = ["B", "KB", "MB", "GB"];
    const i = Math.min(Math.floor(Math.log10(n) / 3), units.length - 1);
    const value = n / Math.pow(1000, i);
    const decimals = i === 0 ? 0 : value < 10 ? 2 : 1;
    return `${value.toFixed(decimals)} ${units[i]}`;
  }

  function formatDate(iso) {
    if (!iso) return "—";
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return iso;
    return d.toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    });
  }

  function setStatus(text, isError) {
    statusEl.textContent = text;
    statusEl.classList.toggle("error", Boolean(isError));
  }

  function renderRows(items) {
    tbody.replaceChildren();
    if (!items.length) {
      const tr = document.createElement("tr");
      tr.className = "empty-row";
      const td = document.createElement("td");
      td.colSpan = 3;
      td.textContent = "Нет объектов с этим префиксом.";
      tr.appendChild(td);
      tbody.appendChild(tr);
      return;
    }

    for (const row of items) {
      const tr = document.createElement("tr");
      const keyTd = document.createElement("td");
      keyTd.className = "key-cell";
      keyTd.textContent = row.key;
      const sizeTd = document.createElement("td");
      sizeTd.className = "num";
      sizeTd.textContent = formatBytes(Number(row.size) || 0);
      const dateTd = document.createElement("td");
      dateTd.textContent = formatDate(row.lastModified);
      tr.append(keyTd, sizeTd, dateTd);
      tbody.appendChild(tr);
    }
  }

  async function loadObjects() {
    const prefix = prefixInput.value.trim();
    const params = new URLSearchParams();
    if (prefix) params.set("prefix", prefix);

    setStatus("Загрузка…", false);
    refreshBtn.disabled = true;

    try {
      const url = `/api/objects${params.toString() ? `?${params}` : ""}`;
      const res = await fetch(url, { headers: { Accept: "application/json" } });
      if (!res.ok) {
        const body = await res.text();
        throw new Error(body || `${res.status} ${res.statusText}`);
      }
      const data = await res.json();
      if (!Array.isArray(data)) {
        throw new Error("Неверный ответ сервера");
      }
      renderRows(data);
      setStatus(`Показано объектов: ${data.length}`, false);
    } catch (e) {
      tbody.replaceChildren();
      const tr = document.createElement("tr");
      tr.className = "empty-row";
      const td = document.createElement("td");
      td.colSpan = 3;
      td.textContent = "Не удалось загрузить список.";
      tr.appendChild(td);
      tbody.appendChild(tr);
      setStatus(e instanceof Error ? e.message : String(e), true);
    } finally {
      refreshBtn.disabled = false;
    }
  }

  refreshBtn.addEventListener("click", () => {
    loadObjects();
  });

  function setUploadStatus(text, isError) {
    uploadStatusEl.textContent = text;
    uploadStatusEl.classList.toggle("error", Boolean(isError));
  }

  function renderUploadLog(items) {
    uploadLogEl.replaceChildren();
    if (!items.length) {
      uploadLogEl.hidden = true;
      return;
    }
    uploadLogEl.hidden = false;
    for (const row of items) {
      const li = document.createElement("li");
      if (row.ok) {
        li.className = "ok";
        li.textContent = `${row.source} → ${row.key}`;
      } else {
        li.className = "err";
        li.textContent = `${row.source}: ${row.error || "ошибка"}`;
      }
      uploadLogEl.appendChild(li);
    }
  }

  async function uploadLocal() {
    setUploadStatus("Загрузка…", false);
    uploadBtn.disabled = true;
    renderUploadLog([]);

    try {
      const res = await fetch("/api/upload-local", {
        method: "POST",
        headers: { Accept: "application/json" },
      });
      const bodyText = await res.text();
      let data;
      try {
        data = bodyText ? JSON.parse(bodyText) : {};
      } catch {
        throw new Error(bodyText || `${res.status} ${res.statusText}`);
      }
      if (!res.ok) {
        throw new Error(data.error || bodyText || `${res.status} ${res.statusText}`);
      }
      const results = Array.isArray(data.results) ? data.results : [];
      renderUploadLog(results);
      const ok = results.filter((r) => r.ok).length;
      const fail = results.length - ok;
      if (!results.length) {
        setUploadStatus("Нет поддерживаемых изображений в каталоге.", false);
      } else {
        setUploadStatus(`Готово: успешно ${ok}, ошибок ${fail}.`, fail > 0);
      }
      prefixInput.value = "uploads/";
      await loadObjects();
    } catch (e) {
      renderUploadLog([]);
      setUploadStatus(e instanceof Error ? e.message : String(e), true);
    } finally {
      uploadBtn.disabled = false;
    }
  }

  uploadBtn.addEventListener("click", () => {
    uploadLocal();
  });

  function setCopyStatus(text, isError) {
    copyStatusEl.textContent = text;
    copyStatusEl.classList.toggle("error", Boolean(isError));
  }

  async function copyPrefix() {
    const from = copyFromInput.value.trim();
    const to = copyToInput.value.trim();
    if (!from || !to) {
      setCopyStatus("Укажите префиксы «Откуда» и «Куда».", true);
      return;
    }

    setCopyStatus("Копирование…", false);
    copyBtn.disabled = true;

    try {
      const res = await fetch("/api/copy-prefix", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          from,
          to,
          delete_source: Boolean(copyDeleteInput.checked),
        }),
      });
      const bodyText = await res.text();
      let data;
      try {
        data = bodyText ? JSON.parse(bodyText) : {};
      } catch {
        throw new Error(bodyText || `${res.status} ${res.statusText}`);
      }
      if (!res.ok) {
        throw new Error(data.error || bodyText || `${res.status} ${res.statusText}`);
      }
      const copied = Number(data.copied) || 0;
      const deleted = Number(data.deleted) || 0;
      const errs = Array.isArray(data.errors) ? data.errors : [];
      const errPart = errs.length ? ` Ошибок: ${errs.length}.` : "";
      setCopyStatus(
        `Скопировано: ${copied}.${copyDeleteInput.checked ? ` Удалено исходных: ${deleted}.` : ""}${errPart}`,
        errs.length > 0,
      );
      prefixInput.value = to.endsWith("/") ? to : `${to}/`;
      await loadObjects();
    } catch (e) {
      setCopyStatus(e instanceof Error ? e.message : String(e), true);
    } finally {
      copyBtn.disabled = false;
    }
  }

  copyBtn.addEventListener("click", () => {
    copyPrefix();
  });

  prefixInput.addEventListener("keydown", (ev) => {
    if (ev.key === "Enter") {
      ev.preventDefault();
      loadObjects();
    }
  });

  document.addEventListener("DOMContentLoaded", () => {
    loadObjects();
  });
})();
