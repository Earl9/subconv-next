const state = {
  config: null,
  status: null,
  collectedNodes: null,
  collectedNodeWarnings: [],
  collectedNodeErrors: [],
  nodeCatalogLoading: false,
  nodeCatalogError: "",
  expandedSubscriptionEditors: new Set(),
  subscriptionNodeSearch: {},
  nodeFilter: "",
  protocolFilter: "",
  tagFilter: "",
  revealedSubscriptionUrls: new Set(),
};

const builtinGroupCatalog = [
  "节点选择",
  "自动选择",
  "故障转移",
  "香港",
  "日本",
  "美国",
  "新加坡",
  "台湾",
  "韩国",
  "德国",
  "英国",
  "法国",
  "加拿大",
  "澳大利亚",
  "AI",
  "流媒体",
  "Telegram",
  "GitHub",
  "Microsoft",
  "Apple",
  "国内直连",
  "漏网之鱼",
  "DIRECT",
  "REJECT",
];

const RULESET_PRESETS = [
  { key: "category-ads-all", label: "广告拦截", category: "核心组", description: "拦截广告和追踪器", policy: "REJECT" },
  { key: "private", label: "私有网络", category: "核心组", description: "局域网和私有地址直连", policy: "DIRECT" },
  { key: "cn", label: "国内服务", category: "核心组", description: "国内域名与服务直连", policy: "国内直连" },
  { key: "geolocation-!cn", label: "非中国", category: "核心组", description: "非中国域名走代理", policy: "节点选择" },
  { key: "__final_policy__", label: "漏网之鱼", category: "核心组", description: "未匹配到任何规则的流量", policy: "漏网之鱼", special: "final_policy" },
  { key: "openai", label: "AI 服务", category: "常用服务", description: "OpenAI、Claude 等", policy: "AI" },
  { key: "youtube", label: "油管视频", category: "常用服务", description: "YouTube、YouTube Music", policy: "流媒体" },
  { key: "google", label: "谷歌服务", category: "常用服务", description: "Google 搜索、Gmail、Drive", policy: "节点选择" },
  { key: "github", label: "GitHub", category: "技术服务", description: "GitHub 与 GitLab", policy: "GitHub" },
  { key: "telegram", label: "Telegram", category: "社交通讯", description: "Telegram 域名与 IP", policy: "Telegram" },
  { key: "microsoft", label: "微软服务", category: "常用服务", description: "Microsoft 365、Bing、Azure", policy: "Microsoft" },
  { key: "apple", label: "苹果服务", category: "常用服务", description: "iCloud、App Store、Apple Music", policy: "Apple" },
  { key: "netflix", label: "奈飞", category: "流媒体", description: "Netflix 流媒体", policy: "流媒体" },
];

document.addEventListener("DOMContentLoaded", () => {
  bindButtons();
  bootstrap();
});

function bindButtons() {
  document.getElementById("refresh-btn").addEventListener("click", refreshNow);
  document.getElementById("save-btn").addEventListener("click", saveConfig);
  document.getElementById("config-save-btn").addEventListener("click", saveConfig);
  document.getElementById("copy-sub-url-btn").addEventListener("click", copySubscriptionURL);
  document.getElementById("preview-reload-btn").addEventListener("click", loadPreview);
  document.getElementById("copy-yaml-btn").addEventListener("click", copyPreviewYAML);
  document.getElementById("add-subscription-btn").addEventListener("click", addSubscription);
  document.getElementById("add-inline-btn").addEventListener("click", addInline);
  document.getElementById("add-provider-btn").addEventListener("click", addProvider);
}

async function bootstrap() {
  await Promise.all([loadStatus(), loadConfig(), loadPreview()]);
  renderSubscriptionURL();
  setInterval(loadStatus, 10000);
  setInterval(loadPreview, 20000);
}

async function loadStatus() {
  const response = await fetchJSON("/api/status");
  if (!response) {
    return;
  }
  state.status = response;

  setText("metric-nodes", String(response.node_count || 0));
  setText("metric-sources", String(response.enabled_subscription_count || 0));
  setText("metric-refresh", formatTime(response.last_refresh_at));
}

async function loadConfig() {
  const response = await fetchJSON("/api/config");
  if (!response || !response.ok) {
    showToast("加载配置失败。", true);
    return;
  }

  state.config = response.config;
  invalidateCollectedNodes();
  renderConfig();
}

async function loadPreview() {
  const output = document.getElementById("preview-output");
  try {
    const url = new URL("/sub/mihomo.yaml", window.location.href);
    url.searchParams.set("_ts", String(Date.now()));
    const response = await fetch(url);
    if (!response.ok) {
      output.textContent = "无法加载 YAML 预览。";
      return;
    }
    output.textContent = await response.text();
  } catch (error) {
    console.error(error);
    output.textContent = "无法加载 YAML 预览。";
  }
}

async function refreshNow() {
  const response = await refreshRenderedOutput();
  if (!response || !response.ok) {
    showToast(readAPIError(response) || "刷新失败。", true);
    return;
  }

  invalidateCollectedNodes();
  showToast(`刷新完成，已渲染 ${response.node_count} 个节点。`);
  await Promise.all([loadStatus(), loadPreview()]);
}

async function saveConfig(options = {}) {
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
    showToast(readAPIError(response) || "保存配置失败。", true);
    return;
  }

  state.config = response.config;
  if (!options.preserveNodeCatalog) {
    invalidateCollectedNodes();
  }
  if (options.render !== false) {
    renderConfig();
  }

  const refreshResponse = await refreshRenderedOutput();
  if (!refreshResponse || !refreshResponse.ok) {
    showToast(readAPIError(refreshResponse) || "配置已保存，但渲染输出刷新失败。", true);
    await Promise.all([loadStatus(), loadPreview()]);
    return;
  }

  showToast(options.toastMessage || "配置已保存并重新生成输出。");
  await Promise.all([loadStatus(), loadPreview()]);
}

function renderSubscriptionURL() {
  const url = subscriptionURL();
  const input = document.getElementById("sub-link-output");
  const open = document.getElementById("open-sub-url-btn");
  input.value = url;
  open.href = url;
}

function subscriptionURL() {
  return new URL("/sub/mihomo.yaml", window.location.href).toString();
}

async function copySubscriptionURL() {
  try {
    await copyText(subscriptionURL());
    showToast("订阅链接已复制。");
  } catch (error) {
    console.error(error);
    showToast("复制订阅链接失败，请手动复制。", true);
  }
}

function renderConfig() {
  renderRulesetCatalog();
  renderAdditionalRules();
  renderRuleProviders();
  renderSubscriptions();
  renderInlineSources();
  renderDNSFields();
}

async function refreshRenderedOutput() {
  return fetchJSON("/api/refresh", {
    method: "POST",
  });
}

function renderRulesetCatalog() {
  const container = document.getElementById("ruleset-catalog");
  container.innerHTML = "";

  const grouped = new Map();
  RULESET_PRESETS.forEach((preset) => {
    if (!grouped.has(preset.category)) {
      grouped.set(preset.category, []);
    }
    grouped.get(preset.category).push(preset);
  });

  grouped.forEach((presets, category) => {
    const block = document.createElement("div");
    block.className = "ruleset-category";

    const enabledCount = presets.filter((preset) => isPresetEnabled(preset)).length;

    const title = document.createElement("div");
    title.className = "ruleset-category-title";
    title.innerHTML = `<strong>${category}</strong><span>${enabledCount}/${presets.length} 已启用</span>`;
    block.appendChild(title);

    const grid = document.createElement("div");
    grid.className = "preset-grid";

    presets.forEach((preset) => {
      const active = isPresetEnabled(preset);
      const card = document.createElement("div");
      card.className = "preset-card";
      if (active) {
        card.classList.add("active");
      }

      const top = document.createElement("div");
      top.className = "preset-head";
      const name = document.createElement("strong");
      name.textContent = preset.label;
      top.appendChild(name);

      if (isSpecialPreset(preset)) {
        const badge = document.createElement("span");
        badge.className = "hint-chip";
        badge.textContent = "固定兜底";
        top.appendChild(badge);
      } else {
        const toggle = actionButton(active ? "已选" : "选择", active ? "primary" : "subtle", async () => {
          await togglePreset(preset);
        });
        top.appendChild(toggle);
      }

      const desc = document.createElement("p");
      desc.className = "field-hint";
      desc.textContent = preset.description || "";

      card.append(top, desc);
      grid.appendChild(card);
    });

    block.appendChild(grid);
    container.appendChild(block);
  });
}

function renderCustomGroups() {
  const list = document.getElementById("custom-groups-list");
  list.innerHTML = "";

  state.config.render.custom_proxy_groups = state.config.render.custom_proxy_groups || [];
  const groups = state.config.render.custom_proxy_groups;

  groups.forEach((group, index) => {
    const card = document.createElement("div");
    card.className = "group-card";

    const head = document.createElement("div");
    head.className = "group-head";
    head.innerHTML = `<strong>代理组 ${index + 1}</strong>`;
    head.appendChild(actionButton("删除", "ghost", () => {
      groups.splice(index, 1);
      renderCustomGroups();
    }));

    const grid = document.createElement("div");
    grid.className = "sub-grid";
    grid.append(
      textField("名称", group.name || "", (value) => {
        group.name = value;
      }),
      selectField("类型", group.type || "select", ["select", "url-test", "fallback", "relay"], (value) => {
        group.type = value;
        renderCustomGroups();
      }),
      toggleField("启用", Boolean(group.enabled), (value) => {
        group.enabled = value;
      }),
    );

    if (group.type === "url-test" || group.type === "fallback") {
      grid.append(
        textField("健康检查 URL", group.url || "https://www.gstatic.com/generate_204", (value) => {
          group.url = value;
        }),
        numberField("检测间隔", group.interval || 300, (value) => {
          group.interval = value;
        }),
      );
    }

    const memberField = document.createElement("div");
    memberField.className = "field full-span";
    const label = document.createElement("label");
    label.textContent = "成员";
    const selected = document.createElement("div");
    selected.className = "sub-url-preview";
    selected.textContent = (group.members || []).length ? group.members.join("\n") : "尚未选择成员";
    memberField.append(label, selected, memberPicker(group, index));

    card.append(head, grid, memberField);
    list.appendChild(card);
  });

  if (!groups.length) {
    list.appendChild(emptyState("还没有配置自定义代理组。"));
  }
}

function renderNodeFilters() {
  const protocolSelect = document.getElementById("node-protocol-filter");
  const tagSelect = document.getElementById("node-tag-filter");
  const summary = document.getElementById("node-filter-summary");

  const nodeDetails = statusNodeDetails();
  const filtered = filteredNodeDetails();
  const protocols = countValues(nodeDetails.map((node) => node.protocol).filter(Boolean));
  const tags = countValues(nodeDetails.flatMap((node) => node.tags || []));

  fillSelect(protocolSelect, "全部协议", protocols.map((item) => ({ value: item.value, label: `${item.value} (${item.count})` })), state.protocolFilter);
  fillSelect(tagSelect, "全部地区", tags.map((item) => ({ value: item.value, label: `${item.value} (${item.count})` })), state.tagFilter);

  if (summary) {
    summary.textContent = `当前命中 ${filtered.length} / ${nodeDetails.length} 个节点。`;
  }
}

function renderServiceFields() {
  const container = document.getElementById("service-fields");
  const service = state.config.service || {};

  const fields = [
    textField("监听地址", service.listen_addr || "", (value) => {
      state.config.service.listen_addr = value;
    }),
    numberField("监听端口", service.listen_port || 9876, (value) => {
      state.config.service.listen_port = value;
    }),
    selectField("模板", service.template || "standard", ["lite", "standard", "full"], (value) => {
      state.config.service.template = value;
    }),
    textField("输出路径", service.output_path || "", (value) => {
      state.config.service.output_path = value;
    }),
    textField("缓存目录", service.cache_dir || "", (value) => {
      state.config.service.cache_dir = value;
    }),
    numberField("刷新间隔", service.refresh_interval || 3600, (value) => {
      state.config.service.refresh_interval = value;
    }),
    numberField("拉取超时", service.fetch_timeout_seconds || 15, (value) => {
      state.config.service.fetch_timeout_seconds = value;
    }),
    numberField("单订阅最大字节", service.max_subscription_bytes || 5242880, (value) => {
      state.config.service.max_subscription_bytes = value;
    }),
  ];

  container.replaceChildren(...fields);
}

function renderBaseFields() {
  const container = document.getElementById("render-base-fields");
  const render = ensureRenderConfig();

  const fields = [
    numberField("mixed-port", render.mixed_port || 7890, (value) => {
      render.mixed_port = value;
    }),
    toggleField("allow-lan", Boolean(render.allow_lan), (value) => {
      render.allow_lan = value;
    }),
    selectField("模式", render.mode || "rule", ["rule", "global", "direct"], (value) => {
      render.mode = value;
    }),
    selectField("日志级别", render.log_level || "info", ["debug", "info", "warning", "error", "silent"], (value) => {
      render.log_level = value;
    }),
    toggleField("IPv6", Boolean(render.ipv6), (value) => {
      render.ipv6 = value;
    }),
    toggleField("unified-delay", Boolean(render.unified_delay), (value) => {
      render.unified_delay = value;
    }),
    toggleField("tcp-concurrent", Boolean(render.tcp_concurrent), (value) => {
      render.tcp_concurrent = value;
    }),
    selectField("find-process-mode", render.find_process_mode || "strict", ["off", "strict", "always"], (value) => {
      render.find_process_mode = value;
    }),
    textField("global-client-fingerprint", render.global_client_fingerprint || "chrome", (value) => {
      render.global_client_fingerprint = value;
    }),
  ];

  container.replaceChildren(...fields);
}

function renderDNSFields() {
  const container = document.getElementById("dns-fields");
  container.innerHTML = "";

  const render = ensureRenderConfig();
  const dns = render.dns || {};
  const fallbackFilter = dns.fallback_filter || {};

  container.append(
    textareaField("默认 DNS", joinLines(dns.default_nameserver), (value) => {
      ensureDNSConfig().default_nameserver = parseLines(value);
    }, {
      rows: 4,
      placeholder: "180.76.76.76\n182.254.118.118\n8.8.8.8",
    }),
    textareaField("主 DNS", joinLines(dns.nameserver), (value) => {
      ensureDNSConfig().nameserver = parseLines(value);
    }, {
      rows: 6,
      placeholder: "180.76.76.76\n119.29.29.29\nhttps://dns.alidns.com/dns-query",
    }),
    textareaField("回退 DNS", joinLines(dns.fallback), (value) => {
      ensureDNSConfig().fallback = parseLines(value);
    }, {
      rows: 6,
      placeholder: "https://dns.google/dns-query\ntls://1.0.0.1:853",
    }),
    textareaField("回退过滤 IPCIDR", joinLines(fallbackFilter.ipcidr), (value) => {
      ensureDNSFallbackFilter().ipcidr = parseLines(value);
    }, {
      rows: 4,
      placeholder: "240.0.0.0/4\n0.0.0.0/32",
    }),
    textareaField("回退过滤域名", joinLines(fallbackFilter.domain), (value) => {
      ensureDNSFallbackFilter().domain = parseLines(value);
    }, {
      rows: 4,
      placeholder: "+.google.com\n+.facebook.com",
    }),
    textareaField("Fake-IP 排除", joinLines(dns.fake_ip_filter), (value) => {
      ensureDNSConfig().fake_ip_filter = parseLines(value);
    }, {
      rows: 8,
      placeholder: "*.lan\ntime.windows.com\npool.ntp.org",
    }),
    textareaField("nameserver-policy", formatNameserverPolicy(dns.nameserver_policy), (value) => {
      ensureDNSConfig().nameserver_policy = parseNameserverPolicy(value);
    }, {
      rows: 6,
      placeholder: "rule-set:company = 172.16.0.53, 172.16.0.54\ngeosite:cn = 223.5.5.5, 119.29.29.29\n+.home.arpa = 192.168.1.1",
    }),
  );
}

function renderProfileFields() {
  const container = document.getElementById("profile-fields");
  const profile = state.config.render?.profile || {};

  container.replaceChildren(
    toggleField("记忆策略选择", profile.store_selected ?? true, (value) => {
      ensureProfileConfig().store_selected = value;
    }),
    toggleField("保存 Fake-IP", Boolean(profile.store_fake_ip), (value) => {
      ensureProfileConfig().store_fake_ip = value;
    }),
  );
}

function renderSnifferFields() {
  const container = document.getElementById("sniffer-fields");
  container.innerHTML = "";

  const sniffer = state.config.render?.sniffer || {};
  const http = sniffer.http || {};

  const topGrid = document.createElement("div");
  topGrid.className = "field-grid";
  topGrid.append(
    toggleField("启用嗅探", Boolean(sniffer.enable), (value) => {
      ensureSnifferConfig().enable = value;
    }),
    toggleField("parse-pure-ip", Boolean(sniffer.parse_pure_ip), (value) => {
      ensureSnifferConfig().parse_pure_ip = value;
    }),
    toggleField("HTTP override-destination", Boolean(http.override_destination), (value) => {
      ensureSnifferHTTP().override_destination = value;
    }),
  );

  container.append(
    topGrid,
    textareaField("TLS 端口", joinLines(sniffer.tls?.ports), (value) => {
      ensureSnifferProtocol("tls").ports = parseLines(value);
    }, {
      rows: 3,
      placeholder: "443\n8443",
    }),
    textareaField("HTTP 端口", joinLines(http.ports), (value) => {
      ensureSnifferHTTP().ports = parseLines(value);
    }, {
      rows: 3,
      placeholder: "80\n8080-8880",
    }),
    textareaField("QUIC 端口", joinLines(sniffer.quic?.ports), (value) => {
      ensureSnifferProtocol("quic").ports = parseLines(value);
    }, {
      rows: 3,
      placeholder: "443\n8443",
    }),
  );
}

function renderAdditionalRules() {
  const input = document.getElementById("additional-rules-input");
  const rules = state.config.render?.additional_rules || [];
  input.value = rules.join("\n");
  input.oninput = (event) => {
    state.config.render.additional_rules = event.target.value
      .split("\n")
      .map((item) => item.trim())
      .filter(Boolean);
  };
}

function renderRuleProviders() {
  const list = document.getElementById("providers-list");
  list.innerHTML = "";

  state.config.render.rule_providers = state.config.render.rule_providers || [];
  const providers = state.config.render.rule_providers.filter((provider) => !isPresetManagedProvider(provider));

  providers.forEach((provider, index) => {
    const card = document.createElement("div");
    card.className = "provider-card";

    const head = document.createElement("div");
    head.className = "provider-head";
    head.innerHTML = `<strong>规则提供器 ${index + 1}</strong>`;
    head.appendChild(actionButton("删除", "ghost", () => {
      providers.splice(index, 1);
      renderRuleProviders();
    }));

    const grid = document.createElement("div");
    grid.className = "sub-grid";
    grid.append(
      textField("名称", provider.name || "", (value) => {
        provider.name = value;
      }),
      textField("策略组", provider.policy || "", (value) => {
        provider.policy = value;
      }),
      selectField("类型", provider.type || "http", ["http", "file", "inline"], (value) => {
        provider.type = value;
        renderRuleProviders();
      }),
      selectField("行为", provider.behavior || "classical", ["classical", "domain", "ipcidr"], (value) => {
        provider.behavior = value;
      }),
      selectField("格式", provider.format || "yaml", ["yaml", "text", "mrs"], (value) => {
        provider.format = value;
      }),
      textField("下载代理", provider.proxy || "DIRECT", (value) => {
        provider.proxy = value;
      }),
      numberField("更新间隔", provider.interval || 86400, (value) => {
        provider.interval = value;
      }),
      toggleField("启用", Boolean(provider.enabled), (value) => {
        provider.enabled = value;
      }),
      toggleField("no-resolve", Boolean(provider.no_resolve), (value) => {
        provider.no_resolve = value;
      }),
    );

    if (provider.type === "http" || provider.type === "file") {
      grid.append(
        textField(provider.type === "http" ? "URL" : "文件路径", provider.type === "http" ? (provider.url || "") : (provider.path || ""), (value) => {
          if (provider.type === "http") {
            provider.url = value;
          } else {
            provider.path = value;
          }
        }),
      );
    }

    if (provider.type === "inline") {
      const payloadField = document.createElement("div");
      payloadField.className = "field full-span";
      const label = document.createElement("label");
      label.textContent = "Payload";
      const textarea = document.createElement("textarea");
      textarea.rows = 8;
      textarea.value = (provider.payload || []).join("\n");
      textarea.placeholder = "每行一条规则，例如：\nDOMAIN-SUFFIX,apple.com\nDOMAIN-SUFFIX,icloud.com";
      textarea.addEventListener("input", (event) => {
        provider.payload = event.target.value.split("\n").map((item) => item.trim()).filter(Boolean);
      });
      payloadField.append(label, textarea);
      card.append(head, grid, payloadField);
      list.appendChild(card);
      return;
    }

    card.append(head, grid);
    list.appendChild(card);
  });

  if (!providers.length) {
    list.appendChild(emptyState("还没有配置自定义规则提供器。"));
  }
}

function presetProvider(key) {
  return (state.config.render.rule_providers || []).find((provider) => provider.name === key);
}

function isPresetManagedProvider(provider) {
  return RULESET_PRESETS.some((preset) => preset.key === provider.name);
}

function isSpecialPreset(preset) {
  return preset?.special === "final_policy";
}

function isPresetEnabled(preset) {
  if (isSpecialPreset(preset)) {
    return true;
  }
  return Boolean(presetProvider(preset.key));
}

async function togglePreset(preset) {
  if (isSpecialPreset(preset)) {
    return;
  }
  state.config.render.rule_providers = state.config.render.rule_providers || [];
  const existing = presetProvider(preset.key);
  if (existing) {
    state.config.render.rule_providers = state.config.render.rule_providers.filter((provider) => provider.name !== preset.key);
  } else {
    state.config.render.rule_providers.push({
      name: preset.key,
      type: "http",
      behavior: "classical",
      format: "yaml",
      interval: 86400,
      proxy: "DIRECT",
      policy: preset.policy,
      no_resolve: preset.key === "private",
      enabled: true,
      url: `https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/refs/heads/meta/geo/geosite/classical/${preset.key}.yaml`,
      payload: [],
    });
  }
  renderConfig();
  await saveConfig({
    render: false,
    toastMessage: "规则集已更新。",
  });
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
      <strong>订阅源 ${index + 1}</strong>
      <div class="card-actions"></div>
    `;

    const actions = head.querySelector(".card-actions");
    actions.append(
      actionButton(state.expandedSubscriptionEditors.has(index) ? "收起节点" : "筛选节点", "subtle", async () => {
        await toggleSubscriptionNodeEditor(index);
      }),
      actionButton(state.revealedSubscriptionUrls.has(index) ? "隐藏地址" : "显示地址", "subtle", () => {
        if (state.revealedSubscriptionUrls.has(index)) {
          state.revealedSubscriptionUrls.delete(index);
        } else {
          state.revealedSubscriptionUrls.add(index);
        }
        renderSubscriptions();
      }),
      actionButton("删除", "ghost", () => {
        state.config.subscriptions.splice(index, 1);
        state.revealedSubscriptionUrls.delete(index);
        renderSubscriptions();
      }),
    );

    const grid = document.createElement("div");
    grid.className = "sub-grid";
    grid.append(
      textField("名称", subscription.name || "", (value) => {
        subscription.name = value;
      }),
      textField("User-Agent", subscription.user_agent || "", (value) => {
        subscription.user_agent = value;
      }),
      toggleField("启用", Boolean(subscription.enabled), (value) => {
        subscription.enabled = value;
      }),
      toggleField("跳过 TLS 校验", Boolean(subscription.insecure_skip_verify), (value) => {
        subscription.insecure_skip_verify = value;
      }),
    );

    const urlBlock = document.createElement("div");
    urlBlock.className = "field";
    const label = document.createElement("label");
    label.textContent = "订阅链接";
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

    const filterGrid = document.createElement("div");
    filterGrid.className = "sub-grid";
    filterGrid.append(
      textareaField("包含关键词", joinLines(subscription.include_keywords), (value) => {
        subscription.include_keywords = parseLines(value);
      }, {
        rows: 4,
        fullSpan: false,
        placeholder: "JP\n日本\n香港",
      }),
      textareaField("排除关键词", joinLines(subscription.exclude_keywords), (value) => {
        subscription.exclude_keywords = parseLines(value);
      }, {
        rows: 4,
        fullSpan: false,
        placeholder: "试用\n过期\n倍率",
      }),
    );

    card.append(head, grid, urlBlock, filterGrid);
    if (state.expandedSubscriptionEditors.has(index)) {
      card.appendChild(subscriptionNodeEditor(subscription, index));
    }
    list.appendChild(card);
  });

  if (!subscriptions.length) {
    list.appendChild(emptyState("还没有配置远程订阅。"));
  }
}

function subscriptionNodeEditor(subscription, index) {
  const wrapper = document.createElement("div");
  wrapper.className = "field full-span";

  const header = document.createElement("div");
  header.className = "member-pane-title";
  const title = document.createElement("strong");
  title.textContent = "手动节点筛选";
  const count = document.createElement("span");

  const sourceNodes = subscriptionSourceNodes(subscription);
  const nodes = visibleSubscriptionNodes(subscription);
  const excluded = new Set(subscription.excluded_node_ids || []);
  const checkedCount = sourceNodes.filter((node) => !excluded.has(node.id)).length;
  count.textContent = `${checkedCount}/${sourceNodes.length} 保留`;
  header.append(title, count);
  wrapper.appendChild(header);

  if (state.nodeCatalogLoading) {
    wrapper.appendChild(emptyState("正在加载节点列表..."));
    return wrapper;
  }
  if (state.nodeCatalogError) {
    wrapper.appendChild(emptyState(state.nodeCatalogError));
    return wrapper;
  }
  if (!state.collectedNodes) {
    wrapper.appendChild(emptyState("展开后会自动加载该订阅的节点列表。"));
    return wrapper;
  }
  if (!nodes.length) {
    wrapper.appendChild(emptyState("当前订阅没有可筛选节点。可先检查订阅链接，或点击刷新后再试。"));
    return wrapper;
  }

  const toolbar = document.createElement("div");
  toolbar.className = "sub-grid";
  toolbar.append(
    textField("搜索节点", state.subscriptionNodeSearch[index] || "", (value) => {
      state.subscriptionNodeSearch[index] = value;
      renderSubscriptions();
    }),
  );

  const actionRow = document.createElement("div");
  actionRow.className = "topbar-actions";
  actionRow.append(
    actionButton("全部保留", "subtle", async () => {
      subscription.excluded_node_ids = [];
      renderSubscriptions();
      await saveConfig({
        render: false,
        preserveNodeCatalog: true,
        toastMessage: "节点筛选已更新。",
      });
    }),
    actionButton("全部排除", "ghost", async () => {
      subscription.excluded_node_ids = sourceNodes.map((node) => node.id);
      renderSubscriptions();
      await saveConfig({
        render: false,
        preserveNodeCatalog: true,
        toastMessage: "节点筛选已更新。",
      });
    }),
  );
  wrapper.append(toolbar, actionRow);

  const list = document.createElement("div");
  list.className = "member-list node-checklist";
  nodes.forEach((node) => {
    list.appendChild(subscriptionNodeRow(subscription, node));
  });
  wrapper.appendChild(list);
  return wrapper;
}

function subscriptionNodeRow(subscription, node) {
  const row = document.createElement("label");
  row.className = "member-item node-check-row";

  const checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.checked = !isSubscriptionNodeExcluded(subscription, node.id);
  checkbox.addEventListener("change", async (event) => {
    setSubscriptionNodeExcluded(subscription, node.id, !event.target.checked);
    renderSubscriptions();
    await saveConfig({
      render: false,
      preserveNodeCatalog: true,
      toastMessage: "节点筛选已更新。",
    });
  });

  const meta = document.createElement("div");
  meta.className = "member-meta";
  const title = document.createElement("strong");
  title.textContent = node.name;
  const detail = document.createElement("span");
  detail.textContent = [node.type?.toUpperCase(), node.server, node.tags?.join(" / ")].filter(Boolean).join(" · ");
  meta.append(title, detail);

  row.append(checkbox, meta);
  return row;
}

async function toggleSubscriptionNodeEditor(index) {
  if (state.expandedSubscriptionEditors.has(index)) {
    state.expandedSubscriptionEditors.delete(index);
    renderSubscriptions();
    return;
  }

  state.expandedSubscriptionEditors.add(index);
  renderSubscriptions();
  await ensureCollectedNodesLoaded();
  renderSubscriptions();
}

async function ensureCollectedNodesLoaded(force = false) {
  if (state.nodeCatalogLoading) {
    return;
  }
  if (state.collectedNodes && !force) {
    return;
  }

  state.nodeCatalogLoading = true;
  state.nodeCatalogError = "";
  try {
    const response = await fetchJSON("/api/nodes");
    if (!response || !response.ok) {
      state.nodeCatalogError = readAPIError(response) || "加载节点列表失败。";
      return;
    }
    state.collectedNodes = response.nodes || [];
    state.collectedNodeWarnings = response.warnings || [];
    state.collectedNodeErrors = response.errors || [];
  } finally {
    state.nodeCatalogLoading = false;
  }
}

function visibleSubscriptionNodes(subscription) {
  const keyword = String(findSubscriptionSearch(subscription) || "").trim().toLowerCase();
  return subscriptionSourceNodes(subscription).filter((node) => {
    if (!keyword) {
      return true;
    }
    const haystack = [node.name, node.type, node.server, (node.tags || []).join(" ")].join(" ").toLowerCase();
    return haystack.includes(keyword);
  });
}

function subscriptionSourceNodes(subscription) {
  const sourceName = String(subscription.name || "");
  return (state.collectedNodes || [])
    .filter((node) => node.source_kind === "subscription" && node.source_name === sourceName)
    .sort((a, b) => a.name.localeCompare(b.name, "zh-CN"));
}

function findSubscriptionSearch(subscription) {
  const index = (state.config.subscriptions || []).indexOf(subscription);
  return state.subscriptionNodeSearch[index] || "";
}

function isSubscriptionNodeExcluded(subscription, nodeID) {
  return (subscription.excluded_node_ids || []).includes(nodeID);
}

function setSubscriptionNodeExcluded(subscription, nodeID, excluded) {
  const next = new Set(subscription.excluded_node_ids || []);
  if (excluded) {
    next.add(nodeID);
  } else {
    next.delete(nodeID);
  }
  subscription.excluded_node_ids = Array.from(next);
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
    head.innerHTML = `<strong>手动源 ${index + 1}</strong>`;
    head.appendChild(actionButton("删除", "ghost", () => {
      state.config.inline.splice(index, 1);
      renderInlineSources();
    }));

    const grid = document.createElement("div");
    grid.className = "sub-grid";
    grid.append(
      textField("名称", inline.name || "", (value) => {
        inline.name = value;
      }),
      toggleField("启用", Boolean(inline.enabled), (value) => {
        inline.enabled = value;
      }),
    );

    const content = document.createElement("div");
    content.className = "field";
    const label = document.createElement("label");
    label.textContent = "节点 / YAML 内容";
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
    list.appendChild(emptyState("还没有配置手动导入内容。"));
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
    include_keywords: [],
    exclude_keywords: [],
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

function addProvider() {
  state.config.render.rule_providers = state.config.render.rule_providers || [];
  state.config.render.rule_providers.push({
    name: "",
    type: "http",
    behavior: "classical",
    format: "yaml",
    interval: 86400,
    proxy: "DIRECT",
    policy: "节点选择",
    no_resolve: false,
    enabled: true,
    payload: [],
  });
  renderRuleProviders();
}

function addCustomGroup() {
  state.config.render.custom_proxy_groups = state.config.render.custom_proxy_groups || [];
  state.config.render.custom_proxy_groups.push({
    name: "",
    type: "select",
    members: [],
    url: "https://www.gstatic.com/generate_204",
    interval: 300,
    enabled: true,
  });
  renderCustomGroups();
}

function memberPicker(group, currentIndex) {
  const wrapper = document.createElement("div");
  wrapper.className = "member-selector";

  const availablePane = document.createElement("div");
  availablePane.className = "member-pane";
  const availableTitle = document.createElement("div");
  availableTitle.className = "member-pane-title";
  const availableOptions = availableMemberOptions(currentIndex).filter((option) => !(group.members || []).includes(option));
  availableTitle.innerHTML = `<strong>可选成员</strong><span>${availableOptions.length}</span>`;
  const availableList = document.createElement("div");
  availableList.className = "member-list";

  if (!availableOptions.length) {
    availableList.appendChild(emptyState("当前筛选条件下没有可选成员。"));
  } else {
    availableOptions.forEach((option) => {
      availableList.appendChild(memberRow(option, "添加", "add", () => {
        group.members = group.members || [];
        if (!group.members.includes(option)) {
          group.members.push(option);
        }
        renderCustomGroups();
      }));
    });
  }

  availablePane.append(availableTitle, availableList);

  const selectedPane = document.createElement("div");
  selectedPane.className = "member-pane";
  const selectedTitle = document.createElement("div");
  selectedTitle.className = "member-pane-title";
  const selectedMembers = group.members || [];
  selectedTitle.innerHTML = `<strong>已选成员</strong><span>${selectedMembers.length}</span>`;
  const selectedList = document.createElement("div");
  selectedList.className = "member-list";

  if (!selectedMembers.length) {
    selectedList.appendChild(emptyState("尚未选择成员。"));
  } else {
    selectedMembers.forEach((option) => {
      selectedList.appendChild(memberRow(option, "移除", "remove", () => {
        group.members = (group.members || []).filter((item) => item !== option);
        renderCustomGroups();
      }));
    });
  }

  selectedPane.append(selectedTitle, selectedList);
  wrapper.append(availablePane, selectedPane);
  return wrapper;
}

function availableMemberOptions(currentIndex) {
  const names = new Set();
  builtinGroupCatalog.forEach((item) => names.add(item));
  filteredNodeDetails().forEach((item) => names.add(item.name));
  (state.config.render.custom_proxy_groups || []).forEach((group, index) => {
    if (index !== currentIndex && group.name) {
      names.add(group.name);
    }
  });
  return Array.from(names).filter(Boolean).sort((a, b) => a.localeCompare(b, "zh-CN"));
}

function ensureRenderConfig() {
  if (!state.config.render) {
    state.config.render = {};
  }
  return state.config.render;
}

function invalidateCollectedNodes() {
  state.collectedNodes = null;
  state.collectedNodeWarnings = [];
  state.collectedNodeErrors = [];
  state.nodeCatalogLoading = false;
  state.nodeCatalogError = "";
}

function ensureDNSConfig() {
  const render = ensureRenderConfig();
  if (!render.dns) {
    render.dns = {
      enable: render.dns_enabled ?? true,
      enhanced_mode: render.enhanced_mode || "fake-ip",
    };
  }
  return render.dns;
}

function ensureDNSFallbackFilter() {
  const dns = ensureDNSConfig();
  if (!dns.fallback_filter) {
    dns.fallback_filter = {
      geoip: true,
      ipcidr: [],
      domain: [],
    };
  }
  return dns.fallback_filter;
}

function ensureProfileConfig() {
  const render = ensureRenderConfig();
  if (!render.profile) {
    render.profile = {
      store_selected: true,
      store_fake_ip: false,
    };
  }
  return render.profile;
}

function ensureSnifferConfig() {
  const render = ensureRenderConfig();
  if (!render.sniffer) {
    render.sniffer = {
      enable: true,
      parse_pure_ip: true,
    };
  }
  return render.sniffer;
}

function ensureSnifferProtocol(key) {
  const sniffer = ensureSnifferConfig();
  if (!sniffer[key]) {
    sniffer[key] = {
      ports: [],
    };
  }
  return sniffer[key];
}

function ensureSnifferHTTP() {
  const http = ensureSnifferProtocol("http");
  if (typeof http.override_destination !== "boolean") {
    http.override_destination = true;
  }
  return http;
}

function memberRow(name, actionText, actionKind, onClick) {
  const row = document.createElement("div");
  row.className = "member-item";

  const meta = document.createElement("div");
  meta.className = "member-meta";
  const title = document.createElement("strong");
  title.textContent = name;
  const type = document.createElement("span");
  type.textContent = memberTypeLabel(name);
  meta.append(title, type);

  const button = actionButton(actionText, actionKind === "add" ? "subtle" : "ghost", onClick);
  row.append(meta, button);
  return row;
}

function memberTypeLabel(name) {
  if ((state.status?.node_names || []).includes(name)) {
    return inferNodeProtocol(name).toUpperCase();
  }
  return "分组";
}

function fillSelect(select, firstLabel, values, currentValue) {
  select.innerHTML = "";
  const first = document.createElement("option");
  first.value = "";
  first.textContent = firstLabel;
  select.appendChild(first);

  values.forEach((value) => {
    const option = document.createElement("option");
    const actualValue = typeof value === "string" ? value : value.value;
    const actualLabel = typeof value === "string" ? value : value.label;
    option.value = actualValue;
    option.textContent = actualLabel;
    if (actualValue === currentValue) {
      option.selected = true;
    }
    select.appendChild(option);
  });
}

function statusNodeDetails() {
  return (state.status?.node_names || []).map((name) => ({
    name,
    protocol: inferNodeProtocol(name),
    tags: inferNodeTags(name),
  }));
}

function filteredNodeDetails() {
  const keyword = state.nodeFilter.trim().toLowerCase();
  return statusNodeDetails().filter((node) => {
    if (keyword && !node.name.toLowerCase().includes(keyword)) {
      return false;
    }
    if (state.protocolFilter && node.protocol !== state.protocolFilter) {
      return false;
    }
    if (state.tagFilter && !(node.tags || []).includes(state.tagFilter)) {
      return false;
    }
    return true;
  });
}

function countValues(values) {
  const counts = new Map();
  values.forEach((value) => {
    counts.set(value, (counts.get(value) || 0) + 1);
  });
  return Array.from(counts.entries())
    .map(([value, count]) => ({ value, count }))
    .sort((a, b) => a.value.localeCompare(b.value, "zh-CN"));
}

function inferNodeProtocol(name) {
  const match = /^\[([^\]]+)\]/.exec(name);
  return match ? match[1].toLowerCase() : "other";
}

function inferNodeTags(name) {
  const tags = [];
  const upper = name.toUpperCase();
  if (upper.includes("JP") || upper.includes("日本")) tags.push("日本");
  if (upper.includes("US") || upper.includes("美国")) tags.push("美国");
  if (upper.includes("HK") || upper.includes("香港")) tags.push("香港");
  if (upper.includes("SG") || upper.includes("新加坡")) tags.push("新加坡");
  if (upper.includes("TW") || upper.includes("台湾")) tags.push("台湾");
  if (upper.includes("KR") || upper.includes("韩国")) tags.push("韩国");
  if (upper.includes("NL")) tags.push("荷兰");
  if (upper.includes("RU")) tags.push("俄罗斯");
  if (upper.includes("GB") || upper.includes("UK")) tags.push("英国");
  return tags;
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

function textareaField(labelText, value, onChange, options = {}) {
  const field = document.createElement("div");
  field.className = options.fullSpan === false ? "field" : "field full-span";
  const label = document.createElement("label");
  label.textContent = labelText;
  const textarea = document.createElement("textarea");
  textarea.rows = options.rows || 6;
  textarea.value = value || "";
  if (options.placeholder) {
    textarea.placeholder = options.placeholder;
  }
  textarea.addEventListener("input", (event) => onChange(event.target.value));
  field.append(label, textarea);
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

  const row = document.createElement("label");
  row.className = "switch";
  const input = document.createElement("input");
  input.type = "checkbox";
  input.checked = checked;
  input.addEventListener("change", (event) => onChange(event.target.checked));

  const track = document.createElement("span");
  track.className = "switch-track";
  const text = document.createElement("span");
  text.className = "switch-text";
  text.textContent = checked ? "开启" : "关闭";
  input.addEventListener("change", () => {
    text.textContent = input.checked ? "开启" : "关闭";
  });

  row.append(input, track, text);
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

async function copyPreviewYAML() {
  const content = document.getElementById("preview-output").textContent || "";
  if (!content || content.includes("无法加载")) {
    showToast("当前没有可复制的 YAML 内容。", true);
    return;
  }

  try {
    await copyText(content);
    showToast("YAML 已复制到剪贴板。");
  } catch (error) {
    console.error(error);
    showToast("复制失败，请手动复制。", true);
  }
}

async function copyText(value) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(value);
    return;
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "readonly");
  textarea.style.position = "fixed";
  textarea.style.top = "-9999px";
  textarea.style.left = "-9999px";
  textarea.style.opacity = "0";
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  textarea.setSelectionRange(0, textarea.value.length);

  try {
    const copied = document.execCommand("copy");
    if (!copied) {
      throw new Error("execCommand(copy) returned false");
    }
  } finally {
    document.body.removeChild(textarea);
  }
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

function joinLines(values) {
  return (values || []).join("\n");
}

function parseLines(value) {
  return (value || "")
    .split("\n")
    .map((item) => item.trim())
    .filter(Boolean);
}

function formatNameserverPolicy(policy) {
  if (!policy) {
    return "";
  }

  return Object.keys(policy)
    .sort((a, b) => a.localeCompare(b, "zh-CN"))
    .map((key) => `${key} = ${(policy[key] || []).join(", ")}`)
    .join("\n");
}

function parseNameserverPolicy(value) {
  const policy = {};
  parseLines(value).forEach((line) => {
    const index = line.indexOf("=");
    if (index === -1) {
      return;
    }
    const key = line.slice(0, index).trim();
    const values = line
      .slice(index + 1)
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
    if (key && values.length) {
      policy[key] = values;
    }
  });
  return policy;
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
    return "从未";
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
    showToast("请求失败，请检查守护进程。", true);
    return null;
  }
}
