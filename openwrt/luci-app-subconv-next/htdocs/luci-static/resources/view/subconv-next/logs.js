'use strict';
'require subconv-next.api as api';
'require dom';
'require ui';
'require view';

var lineOptions = [ 100, 200, 500, 1000 ];

function assertSuccess(result) {
	if (!result || result.success !== true)
		throw new Error(result && result.message ? result.message : _('RPC returned an error'));

	return result;
}

function loadLogs() {
	return api.settle(api.logs.read('service', 100));
}

function resultValue(result) {
	return result && result.ok && result.value ? result.value : null;
}

function option(value, label, selected) {
	var attrs = { 'value': String(value) };

	if (selected)
		attrs.selected = 'selected';

	return E('option', attrs, [ label ]);
}

function autoRefreshControl(viewInstance, root) {
	var inputAttrs = {
		'type': 'checkbox',
		'aria-label': _('自动刷新'),
		'change': function(event) {
			setAutoRefresh(viewInstance, root, event.currentTarget.checked);
		}
	};

	if (viewInstance._autoRefresh)
		inputAttrs.checked = 'checked';

	return E('div', { 'class': 'scn-logs-auto' }, [
		E('span', {}, [ _('自动刷新') ]),
		E('span', { 'id': 'scn-logs-auto-state', 'class': 'scn-logs-auto-state' }, [
			viewInstance._autoRefresh ? _('开启') : _('关闭')
		]),
		E('label', { 'class': 'scn-logs-switch' }, [
			E('input', inputAttrs),
			E('span', { 'class': 'scn-logs-switch-track', 'aria-hidden': 'true' })
		])
	]);
}

function renderLogsContent(viewInstance, root, state) {
	return [
		E('h2', {}, [ _('SubConv Next 日志') ]),
		E('section', { 'class': 'scn-logs-section' }, [
			E('div', { 'class': 'scn-logs-toolbar' }, [
				E('label', { 'class': 'scn-logs-select' }, [
					E('span', {}, [ _('日志来源') ]),
					E('select', {
						'id': 'scn-logs-source',
						'class': 'cbi-input-select',
						'change': function(event) {
							state.source = event.currentTarget.value;
							return refreshLogs(viewInstance, root, state, null, true);
						}
					}, [
						option('service', _('系统服务'), state.source === 'service'),
						option('application', _('应用日志'), state.source === 'application')
					])
				]),
				E('label', { 'class': 'scn-logs-select' }, [
					E('span', {}, [ _('行数') ]),
					E('select', {
						'id': 'scn-logs-lines',
						'class': 'cbi-input-select',
						'change': function(event) {
							state.lines = Number(event.currentTarget.value);
							return refreshLogs(viewInstance, root, state, null, true);
						}
					}, lineOptions.map(function(lines) {
						return option(lines, String(lines), state.lines === lines);
					}))
				]),
				autoRefreshControl(viewInstance, root),
				E('div', { 'class': 'scn-logs-buttons' }, [
					E('button', {
						'class': 'btn cbi-button',
						'data-logs-action': 'refresh',
						'click': function(event) {
							return refreshLogs(viewInstance, root, state, event.currentTarget, false);
						}
					}, [ _('刷新') ]),
					E('button', {
						'class': 'btn cbi-button cbi-button-negative',
						'data-logs-action': 'clear',
						'click': function(event) {
							return clearLogs(viewInstance, root, state, event.currentTarget);
						}
					}, [ _('清空') ]),
					E('button', {
						'class': 'btn cbi-button cbi-button-save',
						'data-logs-action': 'download',
						'click': function(event) {
							return downloadLogs(state, event.currentTarget);
						}
					}, [ _('下载') ])
				])
			]),
			E('pre', {
				'id': 'scn-log-output',
				'class': 'scn-log-output',
				'tabindex': '0',
				'aria-live': 'polite'
			}),
			E('div', { 'id': 'scn-log-meta', 'class': 'scn-log-meta' })
		])
	];
}

function logsResultLines(result) {
	return result && Array.isArray(result.lines) ? result.lines.map(function(line) {
		return String(line);
	}) : [];
}

function updateLogOutput(root, state, preserveScroll) {
	var output = root.querySelector('#scn-log-output');
	var meta = root.querySelector('#scn-log-meta');
	var result = state.result;
	var lines = logsResultLines(result);
	var nearBottom = output.scrollHeight - output.scrollTop - output.clientHeight < 40;
	var message;

	if (result && result.success === true)
		message = lines.length ? lines.join('\n') : _('暂无日志。');
	else
		message = result && result.message ? result.message : _('日志服务不可用。');

	output.textContent = message;
	meta.textContent = _('显示 %d 行 | 更新时间 %s | %s').format(
		lines.length,
		new Date().toLocaleTimeString(),
		result && result.truncated ? _('内容已截断') : _('内容未截断')
	);

	if (!preserveScroll || nearBottom)
		output.scrollTop = output.scrollHeight;
}

function setLogsControlsDisabled(root, disabled) {
	var controls = root.querySelectorAll('[data-logs-action], #scn-logs-source, #scn-logs-lines');

	for (var i = 0; i < controls.length; i++)
		controls[i].disabled = disabled;
}

function refreshLogs(viewInstance, root, state, button, sourceChanged) {
	var originalLabel = button ? button.textContent : '';

	if (viewInstance._logsRefreshing)
		return Promise.resolve();

	viewInstance._logsRefreshing = true;
	setLogsControlsDisabled(root, true);
	if (button) {
		button.classList.add('spinning');
		button.textContent = _('刷新中…');
	}
	if (sourceChanged) {
		state.result = { success: true, lines: [], truncated: false };
		updateLogOutput(root, state, false);
	}

	return api.logs.read(state.source, state.lines).then(assertSuccess).then(function(result) {
		state.result = result;
		updateLogOutput(root, state, true);
	}).catch(function(error) {
		ui.addNotification(null, E('p', {}, [
			_('日志刷新失败：%s').format(error.message || error)
		]), 'danger');
	}).then(function() {
		viewInstance._logsRefreshing = false;
		setLogsControlsDisabled(root, false);
		if (button && button.isConnected) {
			button.classList.remove('spinning');
			button.textContent = originalLabel;
		}
	});
}

function setAutoRefresh(viewInstance, root, enabled) {
	var stateNode = root.querySelector('#scn-logs-auto-state');

	viewInstance._autoRefresh = enabled;
	stateNode.textContent = enabled ? _('开启') : _('关闭');
	if (viewInstance._autoTimer) {
		window.clearInterval(viewInstance._autoTimer);
		viewInstance._autoTimer = null;
	}

	if (enabled) {
		viewInstance._autoTimer = window.setInterval(function() {
			if (!document.hidden && root.isConnected)
				refreshLogs(viewInstance, root, viewInstance._logsState, null, false);
		}, 30000);
	}
}

function clearLogs(viewInstance, root, state) {
	if (state.source === 'service') {
		ui.showModal(_('确认清空显示'), [
			E('p', {}, [ _('仅清空当前页面显示，不会删除 OpenWrt 系统日志。') ]),
			E('div', { 'class': 'right' }, [
				E('button', { 'class': 'btn', 'click': ui.hideModal }, [ _('取消') ]), ' ',
				E('button', { 'class': 'btn cbi-button-negative important', 'click': function() {
					ui.hideModal();
					state.result = { success: true, lines: [], truncated: false };
					updateLogOutput(root, state, false);
					ui.addNotification(null, E('p', {}, [ _('页面日志已清空，系统日志未修改。') ]), 'info');
				} }, [ _('确认清空') ])
			])
		]);
		return Promise.resolve();
	}

	ui.showModal(_('确认清空应用日志'), [
		E('p', {}, [ _('当前应用日志和受控轮转日志将被永久清空。') ]),
		E('div', { 'class': 'right' }, [
			E('button', { 'class': 'btn', 'click': ui.hideModal }, [ _('取消') ]),
			' ',
			E('button', {
				'class': 'btn cbi-button-negative important',
				'click': function() {
					ui.hideModal();
					setLogsControlsDisabled(root, true);
					return api.logs.clear().then(assertSuccess).then(function() {
						ui.addNotification(null, E('p', {}, [ _('应用日志已清空') ]), 'info');
						return refreshLogs(viewInstance, root, state, null, false);
					}).catch(function(error) {
						setLogsControlsDisabled(root, false);
						ui.addNotification(null, E('p', {}, [
							_('清空失败：%s').format(error.message || error)
						]), 'danger');
					});
				}
			}, [ _('确认清空') ])
		])
	]);
}

function downloadFilename() {
	var now = new Date();
	var pad = function(value) { return String(value).padStart(2, '0'); };

	return 'subconv-next-logs-' + now.getFullYear() +
		pad(now.getMonth() + 1) + pad(now.getDate()) + '-' +
		pad(now.getHours()) + pad(now.getMinutes()) + pad(now.getSeconds()) + '.txt';
}

function downloadLogs(state, button) {
	var originalLabel = button.textContent;

	button.disabled = true;
	button.classList.add('spinning');
	button.textContent = _('准备中…');

	return api.logs.download(state.source).then(assertSuccess).then(function(result) {
		var lines = logsResultLines(result);
		var prefix = result.truncated ? '[日志下载内容已按 512 KiB 限制截断]\n' : '';
		var blob = new Blob([ prefix + lines.join('\n') + (lines.length ? '\n' : '') ], { type: 'text/plain;charset=utf-8' });
		var url = window.URL.createObjectURL(blob);
		var link = E('a', {
			'href': url,
			'download': downloadFilename(),
			'style': 'display:none'
		});

		document.body.appendChild(link);
		link.click();
		link.remove();
		window.setTimeout(function() { window.URL.revokeObjectURL(url); }, 0);
	}).catch(function(error) {
		ui.addNotification(null, E('p', {}, [
			_('下载失败：%s').format(error.message || error)
		]), 'danger');
	}).then(function() {
		if (button.isConnected) {
			button.disabled = false;
			button.classList.remove('spinning');
			button.textContent = originalLabel;
		}
	});
}

return view.extend({
	__init__: function() {
		this.super('__init__', arguments);
		this._autoRefresh = false;
		this._autoTimer = null;
		this._logsRefreshing = false;

		if (!document.getElementById('subconv-next-app-css')) {
			document.querySelector('head').appendChild(E('link', {
				'id': 'subconv-next-app-css',
				'rel': 'stylesheet',
				'href': L.resource('subconv-next/app.css')
			}));
		}
		if (!document.getElementById('subconv-next-logs-css')) {
			document.querySelector('head').appendChild(E('link', {
				'id': 'subconv-next-logs-css',
				'rel': 'stylesheet',
				'href': L.resource('subconv-next/logs.css')
			}));
		}
	},

	load: loadLogs,

	render: function(result) {
		var root = E('div', { 'class': 'subconv-next-app' });
		var state = {
			source: 'service',
			lines: 100,
			result: resultValue(result)
		};

		this._logsState = state;
		dom.content(root, renderLogsContent(this, root, state));
		updateLogOutput(root, state, false);
		return root;
	},

	unload: function() {
		if (this._autoTimer) {
			window.clearInterval(this._autoTimer);
			this._autoTimer = null;
		}
	},

	handleSaveApply: null,
	handleSave: null,
	handleReset: null
});
