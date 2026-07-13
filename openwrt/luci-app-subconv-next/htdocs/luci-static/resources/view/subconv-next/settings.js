'use strict';
'require subconv-next.api as api';
'require subconv-next.network-validation as validation';
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

function field(id, label, help, value, type) {
	return E('div', { class: 'scn-field' }, [
		E('label', { class: 'scn-field-label', for: id }, [ label ]),
		E('div', { class: 'scn-field-control' }, [
			E('input', { id: id, class: 'cbi-input-text', type: type || 'text', value: value == null ? '' : String(value) }),
			E('div', { class: 'scn-help' }, [ help ]),
			E('div', { class: 'scn-error', id: id + '-error' })
		])
	]);
}

function showError(root, id, message) {
	root.querySelector('#' + id + '-error').textContent = message || '';
	root.querySelector('#' + id).classList.toggle('cbi-input-invalid', !!message);
}

function validate(root) {
	var values = {
		listen: root.querySelector('#scn-listen').value.trim(),
		port: root.querySelector('#scn-port').value.trim(),
		public_base_url: root.querySelector('#scn-public-url').value.trim(),
		data_dir: root.querySelector('#scn-data-dir').value.trim()
	};
	var valid = true;
	var errors = {
		'scn-listen': values.listen === '' ? _('监听地址不能为空') : (!validation.isListenAddress(values.listen) ? _('请输入有效的 IP 地址或主机名') : ''),
		'scn-port': !validation.isPort(values.port) ? _('端口必须是 1 至 65535 的整数') : '',
		'scn-public-url': !validation.isPublicBaseURL(values.public_base_url) ? _('请输入以 http:// 或 https:// 开头的有效地址') : '',
		'scn-data-dir': values.data_dir === '' ? _('数据目录不能为空') : ''
	};
	Object.keys(errors).forEach(function(id) {
		showError(root, id, errors[id]);
		if (errors[id]) valid = false;
	});
	if (!valid) return null;
	values.port = Number(values.port);
	return values;
}

return view.extend({
	__init__: function() {
		this.super('__init__', arguments);
		addStyle();
	},

	load: function() {
		return api.config.get();
	},

	render: function(config) {
		var self = this;
		this._root = E('div', { class: 'subconv-next-app' }, [
			E('h2', {}, [ _('SubConv Next 设置') ]),
			E('section', { class: 'scn-section scn-form' }, [
				field('scn-listen', _('监听地址'), _('服务绑定的 IP 地址或主机名。'), config.listen, 'text'),
				field('scn-port', _('端口'), _('服务监听端口，范围 1 至 65535。'), config.port, 'number'),
				field('scn-public-url', _('公共访问地址'), _('可留空；填写时必须使用 http:// 或 https://。'), config.public_base_url, 'url'),
				field('scn-data-dir', _('数据目录'), _('保存配置、规则、模板和订阅数据；后端会检查目录范围。'), config.data_dir, 'text')
			]),
			E('section', { class: 'scn-section' }, [
				E('div', { class: 'scn-actions' }, [
					E('button', { class: 'btn cbi-button-save important', click: function(ev) {
						var values = validate(self._root);
						if (!values) { notify(_('请修正表单中的错误'), 'danger'); return; }
						var button = ev.currentTarget;
						button.disabled = true;
						button.classList.add('spinning');
						button.textContent = _('保存中…');
						return api.config.save(values).then(function(result) {
							if (!result || result.success !== true) throw new Error(result && result.message ? result.message : _('保存失败'));
							notify(_('设置已保存'));
							ui.showModal(_('需要重启服务'), [
								E('p', {}, [ _('新设置将在服务重启后完全生效。是否立即重启？') ]),
								E('div', { class: 'right' }, [
									E('button', { class: 'btn', click: ui.hideModal }, [ _('稍后重启') ]), ' ',
									E('button', { class: 'btn cbi-button-action important', click: function() {
										ui.hideModal();
										return api.service.restart().then(function(reply) {
											if (!reply || reply.success !== true) throw new Error(reply && reply.message ? reply.message : _('重启失败'));
											notify(_('服务已重启'));
										}).catch(function(error) { notify(_('重启失败：%s').format(error.message || error), 'danger'); });
									} }, [ _('立即重启') ])
								])
							]);
						}).catch(function(error) {
							notify(_('保存失败：%s').format(error.message || error), 'danger');
						}).then(function() {
							button.disabled = false;
							button.classList.remove('spinning');
							button.textContent = _('保存设置');
						});
					} }, [ _('保存设置') ])
				])
			])
		]);
		return this._root;
	},

	handleSaveApply: null,
	handleSave: null,
	handleReset: null
});
