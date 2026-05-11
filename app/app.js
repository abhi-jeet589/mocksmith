const API = "/admin/mocks";
const PARAM_TYPES = ["string", "int", "uuid"];
const PARAM_TOKEN_RE = /\{([^/{}]+)\}/g;

const form = document.getElementById("mock-form");
const formTitle = document.getElementById("form-title");
const submitBtn = document.getElementById("submit-btn");
const resetBtn = document.getElementById("reset-btn");
const refreshBtn = document.getElementById("refresh-btn");
const formError = document.getElementById("form-error");
const tableBody = document.querySelector("#mocks-table tbody");
const table = document.getElementById("mocks-table");
const emptyState = document.getElementById("empty-state");
const pathParamsSection = document.getElementById("path-params-section");
const pathParamsRows = document.getElementById("path-params-rows");
const headersRows = document.getElementById("headers-rows");
const addHeaderBtn = document.getElementById("add-header-btn");

const fields = {
  id: document.getElementById("mock-id"),
  method: document.getElementById("method"),
  path: document.getElementById("path"),
  statusCode: document.getElementById("statusCode"),
  contentType: document.getElementById("contentType"),
  body: document.getElementById("body"),
};

fields.path.addEventListener("input", () => refreshPathParams());

addHeaderBtn.addEventListener("click", () => addHeaderRow());

headersRows.addEventListener("click", (e) => {
  const btn = e.target.closest("button[data-remove-header]");
  if (!btn) return;
  btn.closest(".header-row")?.remove();
});

function addHeaderRow(name = "", value = "") {
  const row = document.createElement("div");
  row.className = "header-row";
  row.innerHTML = `
    <input type="text" placeholder="Header-Name" data-header-name>
    <input type="text" placeholder="value" data-header-value>
    <button type="button" class="header-remove" data-remove-header aria-label="Remove">&times;</button>
  `;
  row.querySelector("[data-header-name]").value = name;
  row.querySelector("[data-header-value]").value = value;
  headersRows.appendChild(row);
}

function collectHeaders() {
  const out = {};
  headersRows.querySelectorAll(".header-row").forEach((row) => {
    const name = row.querySelector("[data-header-name]").value.trim();
    if (!name) return;
    const value = row.querySelector("[data-header-value]").value;
    out[name] = value;
  });
  return out;
}

function populateHeaders(headers) {
  headersRows.innerHTML = "";
  if (!headers) return;
  for (const [name, value] of Object.entries(headers)) {
    addHeaderRow(name, value);
  }
}

function parseParamNames(path) {
  const names = [];
  const seen = new Set();
  for (const match of String(path).matchAll(PARAM_TOKEN_RE)) {
    const name = match[1];
    if (!seen.has(name)) {
      seen.add(name);
      names.push(name);
    }
  }
  return names;
}

function refreshPathParams(preset) {
  const names = parseParamNames(fields.path.value);

  if (names.length === 0) {
    pathParamsSection.hidden = true;
    pathParamsRows.innerHTML = "";
    return;
  }

  const existing = {};
  pathParamsRows.querySelectorAll("select[data-param-name]").forEach((sel) => {
    existing[sel.dataset.paramName] = sel.value;
  });

  pathParamsRows.innerHTML = "";
  for (const name of names) {
    const row = document.createElement("div");
    row.className = "path-param-row";
    const selected = (preset && preset[name]) || existing[name] || PARAM_TYPES[0];
    row.innerHTML = `
      <span class="path-param-name"><code>${escapeHtml(name)}</code></span>
      <select data-param-name="${escapeAttr(name)}">
        ${PARAM_TYPES.map(
          (t) => `<option value="${t}"${t === selected ? " selected" : ""}>${t}</option>`,
        ).join("")}
      </select>
    `;
    pathParamsRows.appendChild(row);
  }
  pathParamsSection.hidden = false;
}

function collectPathParams() {
  const params = {};
  pathParamsRows.querySelectorAll("select[data-param-name]").forEach((sel) => {
    params[sel.dataset.paramName] = sel.value;
  });
  return params;
}

async function loadMocks() {
  try {
    const res = await fetch(API);
    if (!res.ok) throw new Error(`status ${res.status}`);
    const mocks = await res.json();
    renderTable(mocks);
  } catch (err) {
    showError(`Failed to load mocks: ${err.message}`);
  }
}

function renderTable(mocks) {
  tableBody.innerHTML = "";
  const list = Array.isArray(mocks) ? mocks : [];

  if (list.length === 0) {
    table.style.display = "none";
    emptyState.hidden = false;
    return;
  }

  table.style.display = "";
  emptyState.hidden = true;

  list
    .slice()
    .sort((a, b) => (a.path + a.method).localeCompare(b.path + b.method))
    .forEach((m) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td><span class="method ${escapeAttr(m.method)}">${escapeHtml(m.method)}</span></td>
        <td><span class="path">${escapeHtml(m.path)}</span></td>
        <td><span class="status-code">${escapeHtml(String(m.statusCode))}</span></td>
        <td>
          <span class="row-actions">
            <button class="small" data-action="edit" data-id="${escapeAttr(m.id)}">Edit</button>
            <button class="small danger" data-action="delete" data-id="${escapeAttr(m.id)}">Delete</button>
          </span>
        </td>
      `;
      tableBody.appendChild(tr);
    });
}

tableBody.addEventListener("click", async (e) => {
  const btn = e.target.closest("button[data-action]");
  if (!btn) return;
  const id = btn.dataset.id;
  if (btn.dataset.action === "edit") {
    await loadIntoForm(id);
  } else if (btn.dataset.action === "delete") {
    if (!confirm("Delete this mock?")) return;
    await deleteMock(id);
  }
});

async function loadIntoForm(id) {
  try {
    const res = await fetch(`${API}/${id}`);
    if (!res.ok) throw new Error(`status ${res.status}`);
    const m = await res.json();
    fields.id.value = m.id;
    fields.method.value = m.method;
    fields.path.value = m.path;
    fields.statusCode.value = m.statusCode;
    fields.contentType.value = m.contentType || "";
    fields.body.value = m.body || "";
    refreshPathParams(m.pathParams || {});
    populateHeaders(m.headers);
    formTitle.textContent = "Edit mock";
    submitBtn.textContent = "Update";
    formError.hidden = true;
    form.scrollIntoView({ behavior: "smooth", block: "start" });
  } catch (err) {
    showError(`Failed to load mock: ${err.message}`);
  }
}

async function deleteMock(id) {
  try {
    const res = await fetch(`${API}/${id}`, { method: "DELETE" });
    if (!res.ok && res.status !== 204) throw new Error(`status ${res.status}`);
    if (fields.id.value === id) resetForm();
    await loadMocks();
  } catch (err) {
    showError(`Failed to delete: ${err.message}`);
  }
}

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  formError.hidden = true;

  const pathParams = collectPathParams();
  const headers = collectHeaders();
  const payload = {
    method: fields.method.value,
    path: fields.path.value,
    statusCode: Number(fields.statusCode.value),
    contentType: fields.contentType.value || undefined,
    body: fields.body.value,
  };
  if (Object.keys(pathParams).length > 0) {
    payload.pathParams = pathParams;
  }
  if (Object.keys(headers).length > 0) {
    payload.headers = headers;
  }

  const id = fields.id.value;
  const url = id ? `${API}/${id}` : API;
  const method = id ? "PUT" : "POST";

  try {
    const res = await fetch(url, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || `request failed (${res.status})`);
    }
    resetForm();
    await loadMocks();
  } catch (err) {
    showError(err.message);
  }
});

resetBtn.addEventListener("click", resetForm);
refreshBtn.addEventListener("click", loadMocks);

function resetForm() {
  form.reset();
  fields.id.value = "";
  fields.statusCode.value = 200;
  fields.contentType.value = "application/json";
  refreshPathParams();
  populateHeaders();
  formTitle.textContent = "New mock";
  submitBtn.textContent = "Create";
  formError.hidden = true;
}

function showError(msg) {
  formError.textContent = msg;
  formError.hidden = false;
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  })[c]);
}

function escapeAttr(s) {
  return escapeHtml(s);
}

loadMocks();
