const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

const source = fs.readFileSync(path.join(__dirname, "app.js"), "utf8");

function setup() {
  const requests = [];
  const elements = new Map();
  const context = vm.createContext({
    AbortController, Headers, URL, URLSearchParams,
    document: {
      addEventListener() {},
      getElementById(id) {
        if (!elements.has(id)) {
          elements.set(id, { disabled: false, innerHTML: "", setAttribute() {} });
        }
        return elements.get(id);
      },
    },
    fetch(url, options = {}) {
      return new Promise((resolve) => requests.push({ url, options, resolve }));
    },
  });
  vm.runInContext(source, context);
  vm.runInContext(`
    renderYamlViewer = () => {};
    renderResult = () => {};
    renderNodeEditor = () => {};
    renderConfigNodeSummary = () => {};
    syncSourceFilterOptions = () => {};
    showToast = () => {};
    state.workspaceId = "workspace-a";
  `, context);
  const state = vm.runInContext("state", context);
  return { context, state, requests, run: (code) => vm.runInContext(code, context) };
}

function yamlResponse(text, status = 200) {
  return { ok: status === 200, status, headers: new Headers(), text: async () => text };
}

function nodesResponse(id) {
  return {
    status: 200,
    headers: new Headers(),
    json: async () => ({ ok: true, nodes: [{ id, name: id, enabled: true }], summary: { total: 1 } }),
  };
}

test("views wait for a workspace before requesting data", async () => {
  const app = setup();
  app.state.workspaceId = "";
  await app.run("Promise.all([loadNodes(), loadYamlPreview()])");
  assert.equal(app.requests.length, 0);
});

test("preview coalesces requests and reads the same-origin workspace output", async () => {
  const app = setup();
  app.state.generatedUrl = "https://published.example/s/token/mihomo.yaml";
  const first = app.run("loadYamlPreview()");
  assert.equal(app.run("loadYamlPreview()"), first);
  assert.equal(app.requests.length, 1);
  assert.equal(app.requests[0].url, "/api/preview-yaml?workspace=workspace-a");
  assert.equal(app.requests[0].options.cache, "no-store");
  app.requests[0].resolve(yamlResponse("proxies: []\n"));
  await first;
  assert.equal(app.state.yamlPreview, "proxies: []\n");
  assert.equal(app.state.yamlRequest, null);
});

test("failed preview preserves generated data and allows retry", async () => {
  const app = setup();
  app.state.yamlPreview = "proxies: []\n";
  app.state.generateStatus = "success";
  const first = app.run("loadYamlPreview()");
  app.requests[0].resolve(yamlResponse("server error", 503));
  await first;
  assert.equal(app.state.yamlPreview, "proxies: []\n");
  assert.equal(app.state.generateStatus, "success");
  assert.match(app.state.yamlError, /HTTP 503/);
  const retry = app.run("loadYamlPreview()");
  app.requests[1].resolve(yamlResponse("rules: []\n"));
  await retry;
  assert.equal(app.state.yamlError, "");
  assert.equal(app.state.yamlPreview, "rules: []\n");
});

test("previous workspace preview cannot overwrite the current preview", async () => {
  const app = setup();
  let finishOldBody;
  const first = app.run("loadYamlPreview()");
  app.requests[0].resolve({
    ...yamlResponse(""),
    text: () => new Promise((resolve) => { finishOldBody = resolve; }),
  });
  await new Promise(setImmediate);
  app.run("cancelWorkspaceViewRequests()");
  app.state.workspaceId = "workspace-b";
  const second = app.run("loadYamlPreview()");
  assert.equal(app.requests[0].options.signal.aborted, true);
  app.requests[1].resolve(yamlResponse("current workspace"));
  await second;
  finishOldBody("old workspace");
  await first;
  assert.equal(app.state.yamlPreview, "current workspace");
});

test("node requests coalesce and discard obsolete filter results", async () => {
  const app = setup();
  const first = app.run("loadNodes()");
  assert.equal(app.run("loadNodes()"), first);
  app.state.nodeFilters.q = "updated";
  const second = app.run("loadNodes()");
  assert.equal(app.requests.length, 2);
  assert.equal(app.requests[0].options.signal.aborted, true);
  assert.match(app.requests[1].url, /q=updated/);
  app.requests[1].resolve(nodesResponse("updated"));
  await second;
  app.requests[0].resolve(nodesResponse("original"));
  await first;
  assert.equal(app.state.nodes[0].id, "updated");
  assert.equal(app.state.nodeRequest, null);
});

test("node cache invalidation discards pending responses", async () => {
  const app = setup();
  const first = app.run("loadNodes()");
  app.run("markNodeCacheStale()");
  const second = app.run("loadNodes()");
  assert.equal(app.requests[0].options.signal.aborted, true);
  app.requests[1].resolve(nodesResponse("saved-node"));
  await second;
  app.requests[0].resolve(nodesResponse("old-node"));
  await first;
  assert.equal(app.state.nodes[0].id, "saved-node");
});
