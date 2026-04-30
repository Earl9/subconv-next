'use strict';
'require form';
'require poll';
'require rpc';
'require uci';
'require ui';
'require view';

const serviceName = 'subconv-next';

const callServiceList = rpc.declare({
	object: 'service',
	method: 'list',
	params: [ 'name' ],
	expect: { '': {} }
});

const callInitAction = rpc.declare({
	object: 'luci',
	method: 'setInitAction',
	params: [ 'name', 'action' ],
	expect: { result: false }
});

function serviceRunning(data) {
	var service = data ? data[serviceName] : null;
	var instances = service ? service.instances : null;

	if (!instances)
		return false;

	for (var key in instances) {
		if (instances[key] && instances[key].running)
			return true;
	}

	return false;
}

function loadStatus() {
	return L.resolveDefault(callServiceList(serviceName), {}).then(serviceRunning);
}

function buildWebURL() {
	var host = uci.get(serviceName, 'main', 'host') || '0.0.0.0';
	var port = uci.get(serviceName, 'main', 'port') || '9876';
	var publicBaseURL = uci.get(serviceName, 'main', 'public_base_url') || '';
	var browserHost = window.location.hostname || location.hostname;
	var webHost = host === '0.0.0.0' || host === '::' || host === '' ? browserHost : host;

	return publicBaseURL || window.location.protocol + '//' + webHost + ':' + port + '/';
}

function renderStatus(running) {
	var port = uci.get(serviceName, 'main', 'port') || '9876';
	var dataDir = uci.get(serviceName, 'main', 'data_dir') || '/etc/subconv-next/data';
	var webURL = buildWebURL();
	var badge = running
		? '<span class="label label-success">' + _('Running') + '</span>'
		: '<span class="label label-danger">' + _('Stopped') + '</span>';

	return badge + '<br />' +
		_('Port') + ': <strong>' + port + '</strong><br />' +
		_('Data directory') + ': <code>' + dataDir + '</code><br />' +
		_('Web UI') + ': <a href="' + webURL + '" target="_blank" rel="noopener noreferrer">' + webURL + '</a>';
}

function runAction(action) {
	ui.showModal(null, [
		E('p', { 'class': 'spinning' }, _('Executing %s...').format(action))
	]);

	return callInitAction(serviceName, action).then(function() {
		ui.hideModal();
		location.reload();
	}).catch(function(err) {
		ui.hideModal();
		ui.addNotification(null, E('p', {}, [ _('Service action failed: %s').format(err.message || err) ]), 'danger');
	});
}

return view.extend({
	load: function() {
		return Promise.all([
			uci.load(serviceName),
			loadStatus()
		]);
	},

	render: function(data) {
		var isRunning = data[1];
		var m = new form.Map(serviceName, _('SubConv Next'), _('OpenWrt service wrapper for the SubConv Next Web UI. This LuCI page manages startup and basic runtime configuration only.'));
		var s = m.section(form.NamedSection, 'main', 'subconv-next', _('Service'));
		s.anonymous = true;
		s.addremove = false;

		var status = s.option(form.DummyValue, '_status', _('Status'));
		status.rawhtml = true;
		status.cfgvalue = function() {
			return renderStatus(isRunning);
		};

		var enabled = s.option(form.Flag, 'enabled', _('Enable service'));
		enabled.default = enabled.enabled;
		enabled.rmempty = false;

		var listenHost = s.option(form.Value, 'host', _('Listen address'));
		listenHost.default = '0.0.0.0';
		listenHost.placeholder = '0.0.0.0';
		listenHost.rmempty = false;

		var listenPort = s.option(form.Value, 'port', _('Port'));
		listenPort.datatype = 'port';
		listenPort.default = '9876';
		listenPort.placeholder = '9876';
		listenPort.rmempty = false;

		var dataDirOpt = s.option(form.Value, 'data_dir', _('Data directory'));
		dataDirOpt.default = '/etc/subconv-next/data';
		dataDirOpt.placeholder = '/etc/subconv-next/data';
		dataDirOpt.rmempty = false;

		var publicURL = s.option(form.Value, 'public_base_url', _('Public Base URL'));
		publicURL.placeholder = 'https://subconv.example.com';
		publicURL.rmempty = true;

		var level = s.option(form.ListValue, 'log_level', _('Log level'));
		level.value('debug', _('Debug'));
		level.value('info', _('Info'));
		level.value('warn', _('Warn'));
		level.value('error', _('Error'));
		level.default = 'info';
		level.rmempty = false;

		var actions = s.option(form.DummyValue, '_actions', _('Actions'));
		actions.rawhtml = true;
		actions.cfgvalue = function() {
			return E('div', { 'class': 'cbi-section-actions' }, [
				E('button', {
					'class': 'btn cbi-button cbi-button-apply',
					'click': ui.createHandlerFn(this, runAction, 'start')
				}, [ _('Start') ]),
				' ',
				E('button', {
					'class': 'btn cbi-button cbi-button-reset',
					'click': ui.createHandlerFn(this, runAction, 'stop')
				}, [ _('Stop') ]),
				' ',
				E('button', {
					'class': 'btn cbi-button cbi-button-reload',
					'click': ui.createHandlerFn(this, runAction, 'restart')
				}, [ _('Restart') ]),
				' ',
				E('a', {
					'class': 'btn cbi-button cbi-button-action',
					'href': buildWebURL(),
					'target': '_blank',
					'rel': 'noopener noreferrer'
				}, [ _('Open Web UI') ])
			]).outerHTML;
		};

		poll.add(function() {
			return loadStatus().then(function(running) {
				var node = document.querySelector('[data-name="_status"] .cbi-value-field');
				if (node)
					node.innerHTML = renderStatus(running);
			});
		});

		return m.render();
	},

	handleSaveApply: function(ev, mode) {
		return this.super('handleSaveApply', [ ev, mode ]).then(function() {
			return runAction('restart');
		});
	}
});
