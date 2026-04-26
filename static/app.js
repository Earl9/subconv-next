const state = {
  config: null,
  revealedSubscriptionUrls: new Set(),
};

document.addEventListener("DOMContentLoaded", () => {
  bindButtons();
  bootstrap();
});

function bindButtons() {
  document.getElementById("refresh-btn").addEventListener("click", refreshNow);
  document.getElementById("save-btn").addEventListener("click", saveConfig);
  document.getElementById("config-save-btn").addEventListener("click", saveConfig);
  document.getElementById("config-reload-btn").addEventListener("click", loadConfig);
  document.getElementById("status-reload-btn").addEventListener("click", loadStatus);
  document.getElementById("logs-reload-btn").addEventListener("click", loadLogs);
  document.getElementById("add-subscription-btn").addEventListener("click", addSubscription);
  document.getElementById("add-inline-btn").addEventListener("click", addInline);
}

async function bootstrap() {
  await Promise.all([loadStatus(), loadConfig(), loadLogs()]);
  setInterval(loadStatus, 10000);
  setInterval(loadLogs, 15000);
}

async function loadStatus() {
  const response = await fetchJSON("/api/status");
  if (!response) {
    return;
  }

  setText("metric-running", response.running ? "Online" : "Stopped");
  setText("metric-nodes", String(response.node_count || 0));
  setText("metric-sources", String(response.enabled_subscription_count || 0));
  setText("metric-refresh", formatTime(response.last_refresh_at));
  setText("status-output", response.output_path || "-");
  setText("status-success", formatTime(response.last_success_at));
  setText("status-error", response.last_error || "None");
}

async function loadConfig() {
  const response = await fetchJSON("/api/config");
  if (!response || !response.ok) {
    showToast("Failed to load config.", true);
    return;
  }

  state.config = response.config;
  renderConfig();
}

async function loadLogs() {
  const response = await fetchJSON("/api/logs?tail=200");
  const output = document.getElementById("logs-output");
  if (!response || !response.ok) {
    output.textContent = "Unable to load logs.";
    return;
  }

  output.textContent = response.lines.length ? response.lines.join("\n") : "No log lines yet.";
}

async function refreshNow() {
  const response = await fetchJSON("/api/refresh", {
    method: "POST",
  });
  if (!response || !response.ok) {
    showToast(readAPIError(response) || "Refresh failed.", true);
    return;
  }

  showToast(`Refresh completed. ${response.node_count} node(s) rendered.`);
  await Promise.all([loadStatus(), loadLogs()]);
}

async function saveConfig() {
  if (!state.config) {
    return;
  }

  const response = await fetchJSON("/api/config", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(state.config),
  });

  if (!response || !response.ok) {
    showToast(readAPIError(response) || "Failed to save config.", true);
    return;
  }

  state.config = response.config;
  renderConfig();
  showToast("Configuration saved to /config/config.json.");
  await Promise.all([loadStatus(), loadLogs()]);
}

function renderConfig() {
  renderServiceFields();
  renderSubscriptions();
  renderInlineSources();
}

function renderServiceFields() {
  const container = document.getElementById("service-fields");
  const service = state.config.service || {};

  const fields = [
    textField("Listen Address", service.listen_addr || "", (value) => {
      state.config.service.listen_addr = value;
    }),
    numberField("Listen Port", service.listen_port || 9876, (value) => {
      state.config.service.listen_port = value;
    }),
    selectField("Template", service.template || "standard", ["lite", "standard", "full"], (value) => {
      state.config.service.template = value;
    }),
    textField("Output Path", service.output_path || "", (value) => {
      state.config.service.output_path = value;
    }),
    textField("Cache Dir", service.cache_dir || "", (value) => {
      state.config.service.cache_dir = value;
    }),
    numberField("Refresh Interval", service.refresh_interval || 3600, (value) => {
      state.config.service.refresh_interval = value;
    }),
    numberField("Fetch Timeout", service.fetch_timeout_seconds || 15, (value) => {
      state.config.service.fetch_timeout_seconds = value;
    }),
    numberField("Max Bytes", service.max_subscription_bytes || 5242880, (value) => {
      state.config.service.max_subscription_bytes = value;
    }),
  ];

  container.replaceChildren(...fields);
}

function renderSubscriptions() {
  const list = document.getElementById("subscriptions-list");
  list.innerHTML = "";

  const subscriptions = state.config.subscriptions || [];
  subscriptions.forEach((subscription, index) => {
    const card = document.createElement("div");
    card.className = "sub-card";

    const head = document.createElement("div");
    head.className = "sub-head";
    head.innerHTML = `
      <strong>Subscription ${index + 1}</strong>
      <div class="card-actions"></div>
    `;

    const actions = head.querySelector(".card-actions");
    actions.append(
      actionButton(state.revealedSubscriptionUrls.has(index) ? "Hide URL" : "Reveal URL", "subtle", () => {
        if (state.revealedSubscriptionUrls.has(index)) {
          state.revealedSubscriptionUrls.delete(index);
        } else {
          state.revealedSubscriptionUrls.add(index);
        }
        renderSubscriptions();
      }),
      actionButton("Remove", "ghost", () => {
        state.config.subscriptions.splice(index, 1);
        state.revealedSubscriptionUrls.delete(index);
        renderSubscriptions();
      }),
    );

    const grid = document.createElement("div");
    grid.className = "sub-grid";
    grid.append(
      textField("Name", subscription.name || "", (value) => {
        subscription.name = value;
      }),
      textField("User Agent", subscription.user_agent || "", (value) => {
        subscription.user_agent = value;
      }),
      toggleField("Enabled", Boolean(subscription.enabled), (value) => {
        subscription.enabled = value;
      }),
      toggleField("Skip TLS Verify", Boolean(subscription.insecure_skip_verify), (value) => {
        subscription.insecure_skip_verify = value;
      }),
    );

    const urlBlock = document.createElement("div");
    urlBlock.className = "field";
    const label = document.createElement("label");
    label.textContent = "Subscription URL";
    urlBlock.appendChild(label);

    if (state.revealedSubscriptionUrls.has(index)) {
      const input = document.createElement("textarea");
      input.value = subscription.url || "";
      input.rows = 3;
      input.addEventListener("input", (event) => {
        subscription.url = event.target.value;
      });
      urlBlock.appendChild(input);
    } else {
      const preview = document.createElement("div");
      preview.className = "sub-url-preview";
      preview.textContent = maskSubscriptionURL(subscription.url || "");
      urlBlock.appendChild(preview);
    }

    card.append(head, grid, urlBlock);
    list.appendChild(card);
  });

  if (!subscriptions.length) {
    list.appendChild(emptyState("No subscriptions configured yet."));
  }
}

function renderInlineSources() {
  const list = document.getElementById("inline-list");
  list.innerHTML = "";

  const items = state.config.inline || [];
  items.forEach((inline, index) => {
    const card = document.createElement("div");
    card.className = "inline-card";

    const head = document.createElement("div");
    head.className = "inline-head";
    head.innerHTML = `<strong>Inline Source ${index + 1}</strong>`;
    head.appendChild(actionButton("Remove", "ghost", () => {
      state.config.inline.splice(index, 1);
      renderInlineSources();
    }));

    const grid = document.createElement("div");
    grid.className = "sub-grid";
    grid.append(
      textField("Name", inline.name || "", (value) => {
        inline.name = value;
      }),
      toggleField("Enabled", Boolean(inline.enabled), (value) => {
        inline.enabled = value;
      }),
    );

    const content = document.createElement("div");
    content.className = "field";
    const label = document.createElement("label");
    label.textContent = "Content";
    const textarea = document.createElement("textarea");
    textarea.value = inline.content || "";
    textarea.addEventListener("input", (event) => {
      inline.content = event.target.value;
    });
    content.append(label, textarea);

    card.append(head, grid, content);
    list.appendChild(card);
  });

  if (!items.length) {
    list.appendChild(emptyState("No inline sources configured yet."));
  }
}

function addSubscription() {
  state.config.subscriptions = state.config.subscriptions || [];
  state.config.subscriptions.push({
    name: "",
    enabled: true,
    url: "",
    user_agent: "SubConvNext/0.1",
    insecure_skip_verify: false,
  });
  renderSubscriptions();
}

function addInline() {
  state.config.inline = state.config.inline || [];
  state.config.inline.push({
    name: "",
    enabled: true,
    content: "",
  });
  renderInlineSources();
}

function textField(labelText, value, onChange) {
  const field = document.createElement("div");
  field.className = "field";
  const label = document.createElement("label");
  label.textContent = labelText;
  const input = document.createElement("input");
  input.type = "text";
  input.value = value;
  input.addEventListener("input", (event) => onChange(event.target.value));
  field.append(label, input);
  return field;
}

function numberField(labelText, value, onChange) {
  const field = document.createElement("div");
  field.className = "field";
  const label = document.createElement("label");
  label.textContent = labelText;
  const input = document.createElement("input");
  input.type = "number";
  input.value = value;
  input.addEventListener("input", (event) => onChange(Number(event.target.value)));
  field.append(label, input);
  return field;
}

function selectField(labelText, value, options, onChange) {
  const field = document.createElement("div");
  field.className = "field";
  const label = document.createElement("label");
  label.textContent = labelText;
  const select = document.createElement("select");
  options.forEach((option) => {
    const item = document.createElement("option");
    item.value = option;
    item.textContent = option;
    if (option === value) {
      item.selected = true;
    }
    select.appendChild(item);
  });
  select.addEventListener("change", (event) => onChange(event.target.value));
  field.append(label, select);
  return field;
}

function toggleField(labelText, checked, onChange) {
  const field = document.createElement("div");
  field.className = "field";
  const label = document.createElement("label");
  label.textContent = labelText;

  const row = document.createElement("div");
  row.className = "toggle-row";
  const input = document.createElement("input");
  input.type = "checkbox";
  input.checked = checked;
  input.style.width = "18px";
  input.style.height = "18px";
  input.addEventListener("change", (event) => onChange(event.target.checked));

  const text = document.createElement("span");
  text.textContent = checked ? "Enabled" : "Disabled";
  input.addEventListener("change", () => {
    text.textContent = input.checked ? "Enabled" : "Disabled";
  });

  row.append(input, text);
  field.append(label, row);
  return field;
}

function actionButton(text, kind, onClick) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = `button ${kind}`;
  button.textContent = text;
  button.addEventListener("click", onClick);
  return button;
}

function emptyState(message) {
  const box = document.createElement("div");
  box.className = "sub-url-preview";
  box.textContent = message;
  return box;
}

function maskSubscriptionURL(rawValue) {
  if (!rawValue) {
    return "";
  }

  try {
    const url = new URL(rawValue);
    if (url.username) {
      url.username = "***";
    }
    if (url.password) {
      url.password = "***";
    }
    ["token", "sig", "key", "auth", "password"].forEach((key) => {
      if (url.searchParams.has(key)) {
        url.searchParams.set(key, "***");
      }
    });
    return url.toString();
  } catch {
    return rawValue.replace(/(token=)[^&]+/gi, "$1***");
  }
}

function showToast(message, isError = false) {
  const toast = document.getElementById("toast");
  toast.textContent = message;
  toast.classList.remove("hidden", "error");
  if (isError) {
    toast.classList.add("error");
  }
  window.clearTimeout(showToast.timeout);
  showToast.timeout = window.setTimeout(() => {
    toast.classList.add("hidden");
  }, 3600);
}

function setText(id, value) {
  document.getElementById(id).textContent = value;
}

function formatTime(value) {
  if (!value) {
    return "Never";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function readAPIError(response) {
  return response?.error?.message || "";
}

async function fetchJSON(url, options) {
  try {
    const response = await fetch(url, options);
    return await response.json();
  } catch (error) {
    console.error(error);
    showToast("Request failed. Check daemon logs.", true);
    return null;
  }
}
