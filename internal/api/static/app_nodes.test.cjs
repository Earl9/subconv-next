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

test("published deletion refreshes the result and preview without another generation", async () => {
  const app = setup();
  app.run(`
    const outputRefreshes = [];
    let oldPreviewAborted = false;
    state.yamlPreview = "old published nodes";
    state.resultNodeCount = 3;
    state.yamlRequest = { controller: { abort() { oldPreviewAborted = true; } } };
    renderYamlViewer = () => {};
    renderResult = () => {};
    loadStatus = async () => { outputRefreshes.push("status"); };
    loadPublishedStatus = async () => { outputRefreshes.push("published"); };
    loadAudit = async () => { outputRefreshes.push("audit"); };
    loadYamlPreview = async () => {
      outputRefreshes.push("preview");
      state.yamlPreview = "remaining published node";
    };
  `);
  const pending = app.run('performBulkDeleteNodes([{ id: "deleted" }])');
  respond(app.requests[0], { ok: true, deleted_count: 1, deleted_ids: ["deleted"], published_updated: true, node_count: 2 });
  assert.equal(await pending, true);
  assert.equal(app.requests.length, 1);
  assert.equal(app.run("oldPreviewAborted"), true);
  assert.equal(app.run("state.resultNodeCount"), 2);
  assert.equal(app.run("state.yamlPreview"), "remaining published node");
  assert.deepEqual(Array.from(app.run("outputRefreshes")), ["status", "published", "audit", "preview"]);
  assert.equal(app.run("toasts[0].message"), "已删除 1 个节点。订阅已同步。");
});

test("deleting the last published node clears the preview and explains unavailable output", async () => {
  const app = setup();
  app.run(`
    let previewRefreshes = 0;
    state.yamlPreview = "last node";
    state.resultNodeCount = 1;
    renderYamlViewer = () => {};
    renderResult = () => {};
    loadStatus = async () => {};
    loadPublishedStatus = async () => { state.publishNotice = ""; };
    loadAudit = async () => {};
    loadYamlPreview = async () => { previewRefreshes++; };
  `);
  const pending = app.run('performDeleteCustomNode("last")');
  respond(app.requests[0], { ok: true, deleted_count: 1, deleted_ids: ["last"], published_updated: true, node_count: 0 });
  assert.equal(await pending, true);
  assert.equal(app.run("state.resultNodeCount"), 0);
  assert.equal(app.run("state.yamlPreview"), "");
  assert.equal(app.run("previewRefreshes"), 0);
  assert.equal(app.run("state.publishNotice"), "当前没有可用节点，订阅暂不可下载。");
  assert.equal(app.run("toasts[0].message"), "手动节点已删除。当前订阅已无可用节点。");
});

test("restoring nodes refreshes the original publication and preview without regeneration", async (t) => {
  const cases = [
    { name: "single deleted node", action: 'restoreDeletedNodes(["restored"])', ids: ["restored"] },
    { name: "batch", action: 'restoreDeletedNodes(["restored", "second"])', ids: ["restored", "second"] },
    { name: "all deleted nodes", action: "restoreAllDeletedNodesFromDialog()", ids: ["restored", "second"] },
    { name: "disabled row", action: 'performRestoreNode("restored")', ids: ["restored"] },
    { name: "enable control", action: 'toggleNodeEnabled("restored", true)', ids: ["restored"] },
  ];
  for (const entry of cases) {
    await t.test(entry.name, async () => {
      const app = setup();
      app.elements.set("restore-all-deleted-nodes-btn", { dataset: { restoreDeletedAll: "restored,second" } });
      app.run(`
        const outputRefreshes = [];
        state.resultNodeCount = 0;
        state.publishNotice = "当前没有可用节点，订阅暂不可下载。";
        renderYamlViewer = () => {};
        renderResult = () => {};
        closeDialog = () => {};
        loadStatus = async () => { outputRefreshes.push("status"); };
        loadPublishedStatus = async () => { outputRefreshes.push("published"); };
        loadAudit = async () => { outputRefreshes.push("audit"); };
        loadYamlPreview = async () => {
          outputRefreshes.push("preview");
          state.yamlPreview = "restored published node";
        };
      `);
      const pending = app.run(entry.action);
      assert.equal(app.requests.length, 1);
      assert.match(app.requests[0].url, /^\/api\/nodes\/enable\?/);
      assert.deepEqual(JSON.parse(app.requests[0].options.body).ids, entry.ids);
      respond(app.requests[0], { ok: true, published_updated: true, node_count: entry.ids.length });
      await pending;
      assert.equal(app.requests.length, 1);
      assert.equal(app.run("state.resultNodeCount"), entry.ids.length);
      assert.equal(app.run("state.yamlPreview"), "restored published node");
      assert.equal(app.run("state.publishNotice"), "");
      assert.equal(app.run("refreshCount"), 1);
      assert.deepEqual(Array.from(app.run("outputRefreshes")), ["status", "published", "audit", "preview"]);
      if (entry.name !== "enable control") assert.match(app.run("toasts[0].message"), /订阅已同步。$/);
    });
  }
});

test("failed restoration preserves selection and reports the publication error", async () => {
  const app = setup();
  app.run('state.nodeById.set("deleted", { id: "deleted" }); setNodeSelected("deleted", true, false)');
  const pending = app.run('restoreDeletedNodes(["deleted"])');
  respond(app.requests[0], { ok: false, error: { message: "恢复失败，订阅未更新" } });
  assert.equal(await pending, false);
  assert.equal(app.run("state.selectedNodeIds.size"), 1);
  assert.equal(app.run("state.selectedNodeCache.size"), 1);
  assert.equal(app.run("refreshCount"), 0);
  assert.equal(app.run("toasts[0].message"), "恢复失败，订阅未更新");
  assert.equal(app.run("toasts[0].error"), true);
});

test("node changes from an old workspace cannot modify the current workspace", async (t) => {
  const actions = [
    'performDeleteNode("shared")',
    'performDeleteCustomNode("shared")',
    'performBulkDeleteNodes([{ id: "shared" }])',
    'performRestoreNode("shared")',
    'restoreDeletedNodes(["shared"])',
    'toggleNodeEnabled("shared", true)',
  ];
  for (const action of actions) await t.test(action, async () => {
    const app = setup();
    app.run('state.nodeById.set("shared", { id: "shared", source: { kind: "subscription" } })');
    const pending = app.run(action);
    app.run(`
      state.workspaceId = "new-workspace";
      state.yamlPreview = "new workspace output";
      state.resultNodeCount = 7;
      setNodeSelected("shared", true, false);
      renderYamlViewer = renderResult = () => {};
      loadStatus = loadPublishedStatus = loadAudit = async () => {};
    `);
    respond(app.requests[0], { ok: true, deleted_count: 1, published_updated: true, node_count: 0 });
    await pending;
    assert.equal(app.run("state.yamlPreview"), "new workspace output");
    assert.equal(app.run("state.resultNodeCount"), 7);
    assert.equal(app.run('state.selectedNodeIds.has("shared")'), true);
    assert.equal(app.run("refreshCount"), 0);
    assert.equal(app.run("toasts.length"), 0);
  });
});

test("a completed restore is not overwritten by an older empty-publication refresh", async () => {
  const app = setup();
  app.run(`
    let finishOldAudit;
    let auditLoads = 0;
    renderYamlViewer = renderResult = () => {};
    loadStatus = loadPublishedStatus = async () => {};
    loadAudit = () => ++auditLoads === 1
      ? new Promise((resolve) => { finishOldAudit = resolve; })
      : Promise.resolve();
    loadYamlPreview = async () => { state.yamlPreview = "restored output"; };
  `);
  const deleted = app.run("refreshAfterNodeChange({ published_updated: true, node_count: 0 })");
  await new Promise(setImmediate);
  await app.run("refreshAfterNodeChange({ published_updated: true, node_count: 1 })");
  app.run("finishOldAudit()");
  await deleted;
  assert.equal(app.run("state.resultNodeCount"), 1);
  assert.equal(app.run("state.yamlPreview"), "restored output");
  assert.equal(app.run("state.publishNotice"), "");
});

test("a node-only update does not cancel the pending publication refresh", async () => {
  const app = setup();
  app.run(`
    let finishFirstNodes;
    let nodeLoads = 0;
    refreshNodesAfterGenerate = () => ++nodeLoads === 1
      ? new Promise((resolve) => { finishFirstNodes = resolve; })
      : Promise.resolve();
    renderYamlViewer = renderResult = () => {};
    loadStatus = loadPublishedStatus = loadAudit = async () => {};
    loadYamlPreview = async () => { state.yamlPreview = "updated publication"; };
  `);
  const publication = app.run("refreshAfterNodeChange({ published_updated: true, node_count: 1 })");
  await app.run("refreshAfterNodeChange({ ok: true })");
  app.run("finishFirstNodes()");
  await publication;
  assert.equal(app.run("state.yamlPreview"), "updated publication");
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
