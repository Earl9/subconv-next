const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

function setup() {
  const elements = new Map();
  const requests = [];
  const document = {
    addEventListener() {},
    getElementById(id) { return elements.get(id) || null; },
    createElement(tag) { return new Element(tag); },
    querySelectorAll() { return [...elements.values()].filter((el) => el.getAttribute("aria-invalid")); },
  };
  class Element {
    constructor(tag = "input") {
      this.tagName = tag;
      this.value = "";
      this.textContent = "";
      this.children = [];
      this.attributes = new Map();
      this.listeners = new Map();
      this.classList = {
        classes: new Set(),
        add(value) { this.classes.add(value); },
        remove(value) { this.classes.delete(value); },
        toggle(value, active) { if (active) this.classes.add(value); else this.classes.delete(value); },
        contains(value) { return this.classes.has(value); },
      };
    }
    setAttribute(key, value) { this.attributes.set(key, value); }
    getAttribute(key) { return this.attributes.get(key); }
    removeAttribute(key) { this.attributes.delete(key); }
    appendChild(child) { this.children.push(child); child.parentNode = this; }
    insertAdjacentElement(_, child) { child.parentNode = this; elements.set(child.id, child); }
    remove() { elements.delete(this.id); this.parentNode = null; }
    addEventListener(name, callback) { this.listeners.set(name, callback); }
    focus() { this.focused = true; }
    setSelectionRange(start, end) { this.selection = [start, end]; }
  }
  const values = {
    "edit-index": "", key: "my_rule", label: "My rule", emoji: "", "target-mode": "direct",
    "target-group": "", "source-type": "inline", behavior: "classical", format: "text",
    url: "", path: "", interval: "86400", payload: "DOMAIN-SUFFIX,example.com",
    "enabled": "", "no-resolve": "", "validation": "", "save-custom-rule-btn": "",
  };
  for (const [field, value] of Object.entries(values)) {
    const el = new Element();
    el.id = field === "save-custom-rule-btn" ? field : `custom-rule-${field}`;
    el.value = value;
    elements.set(el.id, el);
  }
  elements.get("custom-rule-enabled").checked = true;
  elements.get("save-custom-rule-btn").textContent = "保存规则";
  for (const id of ["additional-rules-input", "additional-rules-validation", "validate-additional-rules-btn"]) {
    const el = new Element(); el.id = id; elements.set(id, el);
  }
  elements.get("validate-additional-rules-btn").textContent = "校验规则";
  const context = vm.createContext({
    Headers, URL, URLSearchParams, document,
    fetch(url, options) { return new Promise((resolve) => requests.push({ url, options, resolve })); },
  });
  vm.runInContext(fs.readFileSync(path.join(__dirname, "app.js"), "utf8"), context);
  vm.runInContext(`
    state.workspaceId = "rule-test";
    const toasts = [];
    showToast = (message, error) => toasts.push({message, error});
    renderRulePresetControl = renderRuleChips = renderCustomRuleList = resetCustomRuleForm = updateSummary = () => {};
  `, context);
  return { elements, requests, run: (code) => vm.runInContext(code, context), set: (field, value) => { elements.get(`custom-rule-${field}`).value = value; } };
}
function respond(request, body) {
  request.resolve({ status: body.ok ? 200 : 400, headers: new Headers(), json: async () => body });
}

test("invalid required fields, URLs and intervals are explained before requesting validation", async (t) => {
  const cases = [
    ["key", "bad key", "key"],
    ["label", "", "label"],
    ["target-mode", "existing_group", "target-group"],
    ["format", "mrs", "format"],
    ["url", "https://", "url"],
    ["interval", "-1", "interval"],
    ["interval", "1.5", "interval"],
  ];
  for (const [field, value, invalidField] of cases) await t.test(`${field}: ${value}`, async () => {
    const app = setup();
    if (field === "url" || field === "interval") {
      app.set("source-type", "http");
      app.set("url", "https://example.com/rules.yaml");
    }
    if (field === "format") {
      app.set("behavior", "domain");
      app.set("payload", "example.com");
    }
    app.set(field, value);
    assert.equal(await app.run("saveCustomRuleFromDialog()"), false);
    assert.equal(app.requests.length, 0);
    assert.equal(app.run("state.customRules.length"), 0);
    assert.equal(app.elements.get(`custom-rule-${invalidField}`).getAttribute("aria-invalid"), "true");
    assert.equal(app.elements.get("custom-rule-validation").focused, true);
  });
});

test("payload errors preserve the draft and link to the exact original input line", async () => {
  const app = setup();
  const raw = "# comment\n\nIP-CIDR,192.0.2.0/99";
  app.set("payload", raw);
  const pending = app.run("saveCustomRuleFromDialog()");
  assert.equal(app.elements.get("save-custom-rule-btn").disabled, true);
  assert.match(app.requests[0].url, /^\/api\/custom-rules\/validate\?workspace=rule-test$/);
  assert.equal(JSON.parse(app.requests[0].options.body).payload_text, raw);
  assert.equal(app.run("state.customRules.length"), 0);
  respond(app.requests[0], { ok: false, error: { code: "INVALID_CUSTOM_RULE", field: "payload", line: 3, message: "第 3 行：IP 网段格式错误" } });
  assert.equal(await pending, false);
  assert.equal(app.run("state.customRules.length"), 0);
  assert.equal(app.elements.get("custom-rule-payload").value, raw);
  assert.equal(app.elements.get("custom-rule-payload").getAttribute("aria-describedby"), "custom-rule-payload-error");
  assert.match(app.elements.get("custom-rule-payload-error").textContent, /第 3 行/);
  const link = app.elements.get("custom-rule-validation").children[0];
  link.listeners.get("click")({ preventDefault() {} });
  assert.deepEqual(app.elements.get("custom-rule-payload").selection, [11, raw.length]);
  assert.equal(app.elements.get("save-custom-rule-btn").disabled, false);
  assert.equal(app.elements.get("save-custom-rule-btn").textContent, "保存规则");
});

test("saving waits for validation, coalesces repeated clicks and retains valid data", async () => {
  const app = setup();
  const pending = app.run("saveCustomRuleFromDialog()");
  assert.equal(await app.run("saveCustomRuleFromDialog()"), false);
  assert.equal(app.requests.length, 1);
  assert.equal(app.run("state.customRules.length"), 0);
  respond(app.requests[0], { ok: true });
  assert.equal(await pending, true);
  assert.equal(app.run("state.customRules.length"), 1);
  assert.equal(app.run("state.customRules[0].payload[0]"), "DOMAIN-SUFFIX,example.com");
  assert.equal(app.run("state.enabledRules.has('my_rule')"), true);
  assert.equal(app.run("toasts[0].message"), "自定义规则已通过校验并保存。");
  assert.equal(app.elements.get("save-custom-rule-btn").disabled, false);
});

test("validation of an older draft cannot save changed contents", async () => {
  const app = setup();
  const pending = app.run("saveCustomRuleFromDialog()");
  app.set("payload", "IP-CIDR,bad-network");
  respond(app.requests[0], { ok: true });
  assert.equal(await pending, false);
  assert.equal(app.run("state.customRules.length"), 0);
  assert.match(app.run("toasts[0].message"), /内容已修改/);
});

test("failed network validation retains an existing rule and lets the user retry", async () => {
  const app = setup();
  app.run('state.customRules = [{ key: "my_rule", label: "original" }]');
  app.set("edit-index", "0");
  const pending = app.run("saveCustomRuleFromDialog()");
  respond(app.requests[0], { ok: false, error: { message: "无法连接，请重试" } });
  assert.equal(await pending, false);
  assert.equal(app.run("state.customRules[0].label"), "original");
  const retry = app.run("saveCustomRuleFromDialog()");
  respond(app.requests[1], { ok: true });
  assert.equal(await retry, true);
  assert.equal(app.run("state.customRules.length"), 1);
  assert.equal(app.run("state.customRules[0].label"), "My rule");
});

test("an edited rule cannot overwrite a different row moved into its index during validation", async () => {
  const app = setup();
  app.run('state.customRules = [{ key: "my_rule", label: "original" }, { key: "other", label: "keep" }]');
  app.set("edit-index", "0");
  const pending = app.run("saveCustomRuleFromDialog()");
  app.run("state.customRules.splice(0, 1)");
  respond(app.requests[0], { ok: true });
  assert.equal(await pending, false);
  assert.equal(app.run("state.customRules.length"), 1);
  assert.equal(app.run("state.customRules[0].key"), "other");
  assert.match(app.run("toasts[0].message"), /规则列表已修改/);
});

test("pending validation cannot undo a rule disabled from the list", async () => {
  const app = setup();
  app.run('state.customRules = [{ key: "my_rule", label: "original", enabled: true }]');
  app.set("edit-index", "0");
  const pending = app.run("saveCustomRuleFromDialog()");
  app.run("state.customRules[0].enabled = false");
  respond(app.requests[0], { ok: true });
  assert.equal(await pending, false);
  assert.equal(app.run("state.customRules[0].enabled"), false);
  assert.equal(app.run("state.customRules[0].label"), "original");
  assert.match(app.run("toasts[0].message"), /规则列表已修改/);
});

test("direct rules report the raw error line, keep input, and clear the error after correction", async () => {
  const app = setup();
  const input = app.elements.get("additional-rules-input");
  input.value = "# comment\nIP-CIDR,192.0.2.0/99,DIRECT";
  app.run(`
    switchWorkspace = () => {};
    switchRuleSubtab = () => { document.getElementById("additional-rules-input").value = "normalized draft"; };
  `);
  const first = app.run("validateAdditionalRulesInput(true)");
  assert.equal(JSON.parse(app.requests[0].options.body).additional_rules_text, input.value);
  respond(app.requests[0], { ok: false, error: { field: "additional_rules", line: 2, message: "第 2 行：IP 网段格式错误" } });
  assert.equal(await first, false);
  assert.equal(input.value, "# comment\nIP-CIDR,192.0.2.0/99,DIRECT");
  assert.equal(input.getAttribute("aria-invalid"), "true");
  assert.deepEqual(input.selection, [10, input.value.length]);
  assert.match(app.elements.get("additional-rules-validation").textContent, /第 2 行/);
  input.value = "IP-CIDR,192.0.2.0/24,DIRECT";
  const retry = app.run("validateAdditionalRulesInput()");
  respond(app.requests[1], { ok: true });
  assert.equal(await retry, true);
  assert.equal(input.getAttribute("aria-invalid"), undefined);
  assert.equal(app.elements.get("additional-rules-validation").textContent, "直写规则语法校验通过。");
  assert.equal(app.elements.get("validate-additional-rules-btn").disabled, false);
});

test("quoted pasted rules save the same normalized payload that passed validation", async () => {
  const app = setup();
  const raw = '- "DOMAIN-SUFFIX,example.com" # comment';
  app.set("payload", raw);
  const pending = app.run("saveCustomRuleFromDialog()");
  const sent = JSON.parse(app.requests[0].options.body);
  assert.equal(sent.payload_text, raw);
  assert.deepEqual(sent.rule.payload, ["DOMAIN-SUFFIX,example.com"]);
  respond(app.requests[0], { ok: true });
  assert.equal(await pending, true);
  assert.equal(app.run("state.customRules[0].payload[0]"), "DOMAIN-SUFFIX,example.com");
});

test("custom and direct rules preserve literal regex quotes when cleaning pasted scalars", async (t) => {
  const cases = [
    ['- "DOMAIN-SUFFIX,example.com" # comment', "DOMAIN-SUFFIX,example.com"],
    ["-\t'DOMAIN-SUFFIX,example.com' # comment", "DOMAIN-SUFFIX,example.com"],
    ['- "DOMAIN-SUFFIX,example.com" # comment "quoted"', "DOMAIN-SUFFIX,example.com"],
    ["DOMAIN-SUFFIX,example.com\t# comment", "DOMAIN-SUFFIX,example.com"],
    ['DOMAIN-REGEX,^tracker"', 'DOMAIN-REGEX,^tracker"'],
    ['DOMAIN-REGEX,^tracker"#fragment$', 'DOMAIN-REGEX,^tracker"#fragment$'],
    ["-invalid.example", "-invalid.example"],
  ];
  for (const [raw, expected] of cases) await t.test(raw, () => {
    const app = setup();
    app.set("payload", raw);
    assert.equal(app.run('prepareCustomRulePayloadLine(getValue("custom-rule-payload"))'), expected);
    assert.equal(app.run('normalizeAdditionalRuleLine(getValue("custom-rule-payload"))'), expected);
  });
});

test("provider payload cleanup preserves regex literals and routing parameters", async (t) => {
  const cases = [
    ["IP-CIDR,192.0.2.0/24,DIRECT,SRC,NO-RESOLVE", "IP-CIDR,192.0.2.0/24,src,no-resolve"],
    ["GEOIP,CN,proxy,src", "GEOIP,CN,src"],
    ["DOMAIN-REGEX,^foo,DIRECT", "DOMAIN-REGEX,^foo,DIRECT"],
    ["PROCESS-PATH-REGEX,^foo,REJECT", "PROCESS-PATH-REGEX,^foo,REJECT"],
    ["AND,((DOMAIN,example.com),(NETWORK,TCP)),DIRECT", "AND,((DOMAIN,example.com),(NETWORK,TCP))"],
    ["DOMAIN,example.com,DIRECT,REJECT", "DOMAIN,example.com,REJECT"],
  ];
  for (const [raw, expected] of cases) await t.test(raw, () => {
    const app = setup();
    app.set("payload", raw);
    assert.equal(app.run('normalizeCustomRulePayloadLines(getValue("custom-rule-payload"))[0]'), expected);
  });
});
