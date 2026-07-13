'use strict';
'require subconv-next.api as api';
'require dom';
'require ui';
'require view';

function addStyle() {
	if (!document.getElementById('subconv-next-app-css'))
		document.head.appendChild(E('link', { id: 'subconv-next-app-css', rel: 'stylesheet', href: L.resource('subconv-next/app.css') }));
}

function notify(message, type) {
	ui.addNotification(null, E('p', {}, [ message ]), type || 'info');
}

function value(result) {
	return result && result.ok ? result.value : null;
}

function row(label, content) {
	return E('div', { class: 'scn-row' }, [ E('span', { class: 'scn-label' }, [ label ]), E('span', { class: 'scn-value' }, Array.isArray(content) ? content : [ content ]) ]);
}

function statusNode(status) {
	var state = !status ? 'error' : status.running ? 'running' : 'stopped';
	var text = state === 'running' ? _('运行中') : state === 'stopped' ? _('已停止') : _('异常');
	return E('span', { class: 'scn-status scn-status-' + state }, [ E('span', { class: 'scn-status-dot', 'aria-hidden': 'true' }), E('span', {}, [ text ]) ]);
}

function openWebUI(status) {
	if (!status || !status.running) {
		notify(_('服务未运行'), 'danger');
		return;
	}
	if (!status.webUiAvailable || !status.webUiUrl) {
		notify(_('请先在设置中配置 Public Base URL'), 'danger');
		return;
	}
	try {
		var url = new URL(status.webUiUrl);
		if (url.protocol !== 'http:' && url.protocol !== 'https:')
			throw new Error(_('WebUI 地址协议无效'));
		window.open(url.href, '_blank', 'noopener,noreferrer');
	}
	catch (error) {
		notify(_('打开 WebUI 失败：%s').format(error.message || error), 'danger');
	}
}

function confirmAction(instance, action, label) {
	ui.showModal(_('确认操作'), [
		E('p', {}, [ _('确定要%s SubConv Next 服务吗？').format(label) ]),
		E('div', { class: 'right' }, [
			E('button', { class: 'btn', click: ui.hideModal }, [ _('取消') ]), ' ',
			E('button', { class: action === 'stop' ? 'btn cbi-button-negative important' : 'btn cbi-button-action important', click: function() {
				ui.hideModal();
				return runAction(instance, action, label);
			} }, [ _('确认') ])
		])
	]);
}

function runAction(instance, action, label) {
	if (instance._busy)
		return Promise.resolve();
	instance._busy = true;
	instance._busyAction = action;
	instance.renderState();
	return api.service[action]().then(function(result) {
		if (!result || result.success !== true)
			throw new Error(result && result.message ? result.message : _('服务操作失败'));
		notify(_('%s成功').format(label), 'info');
	}).catch(function(error) {
		notify(_('%s失败：%s').format(label, error.message || error), 'danger');
	}).then(function() {
		return instance.refresh();
	}).then(function() {
		instance._busy = false;
		instance._busyAction = null;
		instance.renderState();
	});
}

return view.extend({
	__init__: function() {
		this.super('__init__', arguments);
		this._busy = false;
		this._busyAction = null;
		addStyle();
	},

	load: function() {
		return Promise.all([ api.settle(api.service.status()), api.settle(api.config.get()), api.settle(api.system.getAutostart()) ]);
	},

	render: function(results) {
		this._root = E('div', { class: 'subconv-next-app' });
		this._status = value(results[0]);
		this._config = value(results[1]);
		this._autostart = value(results[2]);
		this.renderState();
		return this._root;
	},

	refresh: function() {
		var self = this;
		return this.load().then(function(results) {
			self._status = value(results[0]);
			self._config = value(results[1]);
			self._autostart = value(results[2]);
			self.renderState();
		});
	},

	renderState: function() {
		var self = this;
		var running = !!(this._status && this._status.running);
		var webUIAvailable = !!(this._status && this._status.webUiAvailable && this._status.webUiUrl);
		var webUITitle = !running ? _('服务未运行') : !webUIAvailable ? _('请先在设置中配置 Public Base URL') : _('在新标签页打开 WebUI');
		var config = this._config || {};
		var autostart = !!(this._autostart && this._autostart.enabled);
		var switchAttrs = { type: 'checkbox', change: function(ev) {
			var enabled = ev.currentTarget.checked;
			ev.currentTarget.disabled = true;
			api.system.setAutostart(enabled).then(function(result) {
				if (!result || result.success !== true) throw new Error(result && result.message ? result.message : _('设置失败'));
				self._autostart = { enabled: enabled };
				notify(enabled ? _('已启用开机启动') : _('已关闭开机启动'));
			}).catch(function(error) {
				notify(_('开机启动设置失败：%s').format(error.message || error), 'danger');
			}).then(function() { self.renderState(); });
		} };
		if (autostart) switchAttrs.checked = 'checked';
		if (this._busy) switchAttrs.disabled = 'disabled';
		var switchInput = E('input', switchAttrs);
		dom.content(this._root, [
			E('h2', {}, [ _('SubConv Next 概览') ]),
			E('section', { class: 'scn-section scn-summary' }, [
				E('div', {}, [ row(_('服务状态'), statusNode(this._status)), row(_('当前版本'), this._status && this._status.version || '-'), row(_('运行时间'), this._status && this._status.uptime || '-') ]),
				E('div', {}, [ row(_('监听地址'), config.listen || '-'), row(_('监听端口'), config.port != null ? String(config.port) : '-'), row(_('开机启动'), E('label', { class: 'scn-switch' }, [ switchInput, E('span', {}, [ autostart ? _('已开启') : _('已关闭') ]) ])) ])
			]),
			E('section', { class: 'scn-section' }, [
				E('div', { class: 'scn-actions' }, [
					E('button', { class: 'btn cbi-button-action', title: webUITitle, disabled: !running || !webUIAvailable || this._busy ? 'disabled' : null, click: function() { openWebUI(self._status); } }, [ E('span', { 'aria-hidden': 'true' }, [ '↗' ]), ' ', _('打开 WebUI') ]),
					E('button', { class: 'btn cbi-button-action', disabled: running || this._busy ? 'disabled' : null, click: function() { confirmAction(self, 'start', _('启动')); } }, [ this._busyAction === 'start' ? _('启动中…') : _('启动') ]),
					E('button', { class: 'btn cbi-button-negative', disabled: !running || this._busy ? 'disabled' : null, click: function() { confirmAction(self, 'stop', _('停止')); } }, [ this._busyAction === 'stop' ? _('停止中…') : _('停止') ]),
					E('button', { class: 'btn cbi-button', disabled: !running || this._busy ? 'disabled' : null, click: function() { confirmAction(self, 'restart', _('重启')); } }, [ this._busyAction === 'restart' ? _('重启中…') : _('重启') ]),
					E('button', { class: 'btn cbi-button', title: _('刷新服务状态'), disabled: this._busy ? 'disabled' : null, click: function(ev) { ev.currentTarget.classList.add('spinning'); return self.refresh().then(function() { ev.currentTarget.classList.remove('spinning'); }); } }, [ _('刷新') ])
				])
			])
		]);
	},

	handleSaveApply: null,
	handleSave: null,
	handleReset: null
});
