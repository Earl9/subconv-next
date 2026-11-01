const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

function setup() {
  const requests = [];
  const elements = new Map();
  const context = vm.createContext({
    Headers, URL, URLSearchParams,
    document: {
      addEventListener() {},
      getElementById(id) {
        if (!elements.has(id)) elements.set(id, { disabled: false, open: false });
        return elements.get(id);
      },
    },
    fetch(url, options) {
      return new Promise((resolve) => requests.push({ url, options, resolve }));
    },
  });
  vm.runInContext(fs.readFileSync(path.join(__dirname, "app.js"), "utf8"), context);
  vm.runInContext(`
    state.workspaceId = "test-workspace";
    let refreshCount = 0;
    const toasts = [];
    refreshNodesAfterGenerate = async () => { refreshCount++; };
    renderNodeTable = () => {};
    showToast = (message, error) => toasts.push({ message, error });
  `, context);
  return { requests, elements, run: (code) => vm.runInContext(code, context) };
}

function respond(request, body) {
  request.resolve({ status: body.ok ? 200 : 404, headers: new Headers(), json: async () => body });
}

test("bulk deletion includes selections from earlier pages in one mixed batch", async () => {
  const app = setup();
  app.run(`
    const first = { id: "subscription", source: { kind: "subscription" } };
    const second = { id: "custom", source: { kind: "custom" } };
    state.nodeById.set(first.id, first);
    setNodeSelected(first.id, true, false);
    state.nodeById = new Map([[second.id, second]]);
    setNodeSelected(second.id, true, false);
  `);
  const pending = app.run('performBulkDeleteNodes(getBulkDeleteTargets("selected"))');
  assert.equal(app.requests.length, 1);
  assert.match(app.requests[0].url, /^\/api\/nodes\/delete\?/);
  assert.deepEqual(JSON.parse(app.requests[0].options.body).ids, ["subscription", "custom"]);
  respond(app.requests[0], { ok: true, deleted_count: 2, deleted_ids: ["subscription", "custom"] });
  assert.equal(await pending, true);
  assert.equal(app.run("state.selectedNodeIds.size"), 0);
  assert.equal(app.run("state.selectedNodeCache.size"), 0);
  assert.equal(app.run("refreshCount"), 1);
  assert.equal(app.run("toasts[0].message"), "已删除 2 个节点。");
});

test("a rejected batch preserves selection and reports the backend error", async () => {
  const app = setup();
  app.run('state.nodeById.set("stale", { id: "stale" }); setNodeSelected("stale", true, false)');
  const pending = app.run('performBulkDeleteNodes(getBulkDeleteTargets("selected"))');
  respond(app.requests[0], { ok: false, error: { message: "请刷新节点列表后重试" } });
  assert.equal(await pending, false);
  assert.equal(app.run("state.selectedNodeIds.size"), 1);
  assert.equal(app.run("state.selectedNodeCache.size"), 1);
  assert.equal(app.run("refreshCount"), 0);
  assert.equal(app.run("toasts[0].message"), "请刷新节点列表后重试");
  assert.equal(app.run("toasts[0].error"), true);
});

test("deleting a disabled node opens deletion confirmation and never restores it", async () => {
  const app = setup();
  app.run(`
    state.editingNode = { id: "disabled", enabled: false, name: "disabled", source: { kind: "subscription" } };
    getValue = () => "disabled";
    openDangerDialog = (config) => { state.confirmDialog = config; };
  `);
  await app.run("handleDeleteCurrentNode()");
  assert.equal(app.requests.length, 0);
  assert.equal(app.run("state.confirmDialog.action"), "delete-node");
  assert.equal(app.run("state.confirmDialog.confirmText"), "删除节点");
  assert.equal(app.run("nodeDeleteSemantics(state.editingNode).inlineConfirmText"), "删除");
});

test("repeated confirmation does not submit the same deletion twice", async () => {
  const app = setup();
  app.run('state.confirmDialog = { action: "bulk-delete-nodes", nodes: [{ id: "node" }] }');
  const pending = app.run("handleDangerDialogConfirm()");
  assert.equal(app.elements.get("danger-dialog-confirm-btn").disabled, true);
  await app.run("handleDangerDialogConfirm()");
  assert.equal(app.requests.length, 1);
  respond(app.requests[0], { ok: true, deleted_count: 1, deleted_ids: ["node"] });
  await pending;
  assert.equal(app.elements.get("danger-dialog-confirm-btn").disabled, false);
  assert.equal(app.run("state.confirmDialog"), null);
});

test("drafts preserve persistent deletion identities and prune removed sources", () => {
  const app = setup();
  app.run(`
    const draftState = {
      deleted_nodes: ["old-connection", "removed-connection"],
      deleted_node_sources: { "old-connection": ["active"], "removed-connection": ["removed"] },
      deleted_node_signatures: { "old-connection": ["stable-identity"], "removed-connection": ["other-identity"] },
    };
    const saved = sanitizeLocalDraftNodeState(draftState);
    const restored = pruneLocalDraftDeletedNodes(saved, { subscriptions: [{ id: "active" }] });
  `);
  assert.equal(app.run('saved.deleted_node_signatures["old-connection"][0]'), "stable-identity");
  assert.equal(app.run('restored.deleted_node_signatures["old-connection"][0]'), "stable-identity");
  assert.equal(app.run('restored.deleted_node_signatures["removed-connection"]'), undefined);
  assert.equal(app.run('draftState.deleted_node_signatures["removed-connection"][0]'), "other-identity");
});
