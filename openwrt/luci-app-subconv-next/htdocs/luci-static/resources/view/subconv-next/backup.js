'use strict';
'require subconv-next.api as api';
'require dom';
'require fs';
'require ui';
'require view';

var MAX_SIZE = 16 * 1024 * 1024;

function addStyle() {
	if (!document.getElementById('subconv-next-app-css'))
		document.head.appendChild(E('link', { id: 'subconv-next-app-css', rel: 'stylesheet', href: L.resource('subconv-next/app.css') }));
}

function notify(message, type) {
	ui.addNotification(null, E('p', {}, [ message ]), type || 'info');
}

function requireSuccess(result, fallback) {
	if (!result || result.success !== true)
		throw new Error(result && result.message ? result.message : fallback);
	return result;
}

function formatSize(bytes) {
	return '%1024mB'.format(bytes || 0);
}

function base64(buffer) {
	var bytes = new Uint8Array(buffer);
	var binary = '';
	for (var i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
	return window.btoa(binary);
}

function downloadBlob(blob, filename) {
	var url = window.URL.createObjectURL(blob);
	var link = E('a', { href: url, download: filename, style: 'display:none' });
	document.body.appendChild(link);
	link.click();
	link.remove();
	window.setTimeout(function() { window.URL.revokeObjectURL(url); }, 0);
}

return view.extend({
	__init__: function() {
		this.super('__init__', arguments);
		this._file = null;
		this._uploadId = null;
		this._checked = null;
		this._busy = false;
		addStyle();
	},

	load: function() {
		return Promise.resolve();
	},

	render: function() {
		var self = this;
		var input = E('input', { type: 'file', accept: '.tar.gz,application/gzip', style: 'display:none', change: function(ev) { self.selectFile(ev.currentTarget.files[0]); } });
		this._fileInfo = E('div');
		this._checkResult = E('div');
		this._checkButton = E('button', { class: 'btn cbi-button', disabled: 'disabled', click: function() { return self.checkBackup(); } }, [ _('检查备份') ]);
		this._restoreButton = E('button', { class: 'btn cbi-button-negative', disabled: 'disabled', click: function() { self.confirmRestore(); } }, [ _('开始恢复') ]);
		var dropzone = E('div', {
			class: 'scn-dropzone', tabindex: '0', role: 'button',
			click: function() { input.click(); },
			keydown: function(ev) { if (ev.key === 'Enter' || ev.key === ' ') { ev.preventDefault(); input.click(); } },
			dragover: function(ev) { ev.preventDefault(); ev.currentTarget.classList.add('is-dragging'); },
			dragleave: function(ev) { ev.currentTarget.classList.remove('is-dragging'); },
			drop: function(ev) { ev.preventDefault(); ev.currentTarget.classList.remove('is-dragging'); self.selectFile(ev.dataTransfer.files[0]); }
		}, [ input, E('strong', {}, [ _('选择或拖入备份文件') ]), E('div', { class: 'scn-muted' }, [ _('仅支持 tar.gz，最大 16 MiB') ]) ]);
		this._root = E('div', { class: 'subconv-next-app' }, [
			E('h2', {}, [ _('备份与恢复') ]),
			E('section', { class: 'scn-section' }, [
				E('h3', {}, [ _('导出备份') ]),
				E('p', { class: 'scn-muted' }, [ _('备份用于迁移或恢复 SubConv Next 数据，包含服务配置和业务数据，不包含日志、缓存及运行时文件。') ]),
				E('div', { class: 'scn-actions' }, [ E('button', { class: 'btn cbi-button-save important', title: _('生成并下载备份'), click: function(ev) { return self.exportBackup(ev.currentTarget); } }, [ _('导出备份') ]) ])
			]),
			E('section', { class: 'scn-section' }, [
				E('h3', {}, [ _('恢复备份') ]),
				E('p', { class: 'scn-muted' }, [ _('上传后会先检查格式、清单、路径和完整性，检查通过后才能恢复。') ]),
				dropzone,
				this._fileInfo,
				this._checkResult,
				E('div', { class: 'scn-actions', style: 'margin-top:12px' }, [ this._checkButton, this._restoreButton ])
			])
		]);
		return this._root;
	},

	selectFile: function(file) {
		var self = this;
		if (this._uploadId) api.backup.cancelUpload(this._uploadId).catch(function() {});
		this._uploadId = null;
		this._checked = null;
		this._file = file || null;
		dom.content(this._checkResult, []);
		if (!file) {
			dom.content(this._fileInfo, []);
			this.updateButtons();
			return;
		}
		var error = !/\.tar\.gz$/i.test(file.name) ? _('文件扩展名必须为 .tar.gz') : file.size <= 0 || file.size > MAX_SIZE ? _('文件大小必须在 16 MiB 以内') : '';
		dom.content(this._fileInfo, E('div', { class: 'scn-file-info' }, [
			E('span', {}, [ file.name ]), E('span', {}, [ formatSize(file.size) ]), error ? E('span', { class: 'scn-error' }, [ error ]) : ''
		]));
		if (error) this._file = null;
		this.updateButtons();
	},

	updateButtons: function() {
		this._checkButton.disabled = this._busy || !this._file;
		this._restoreButton.disabled = this._busy || !this._checked;
	},

	exportBackup: function(button) {
		var cleanupId = null;
		button.disabled = true;
		button.classList.add('spinning');
		button.textContent = _('正在生成…');
		return api.backup.exportArchive().then(function(result) {
			requireSuccess(result, _('导出失败'));
			cleanupId = result.id;
			return fs.read_direct(result.path, 'blob').then(function(blob) { downloadBlob(blob, result.filename); });
		}).then(function() {
			notify(_('备份已导出'));
		}).catch(function(error) {
			notify(_('导出失败：%s').format(error.message || error), 'danger');
		}).then(function() {
			if (cleanupId) api.backup.cleanupExport(cleanupId).catch(function() {});
			button.disabled = false;
			button.classList.remove('spinning');
			button.textContent = _('导出备份');
		});
	},

	checkBackup: function() {
		var self = this;
		var file = this._file;
		this._busy = true;
		this._checked = null;
		this._checkButton.textContent = _('上传并检查中…');
		this.updateButtons();
		return api.backup.beginUpload().then(function(beginResult) {
			var begin = requireSuccess(beginResult, _('无法开始上传'));
			self._uploadId = begin.upload_id;
			var offset = 0;
			var index = 0;
			function sendNext() {
				if (offset >= file.size) return Promise.resolve();
				var end = Math.min(offset + begin.chunk_size, file.size);
				return file.slice(offset, end).arrayBuffer().then(function(buffer) {
					return api.backup.uploadChunk(begin.upload_id, index, base64(buffer));
				}).then(function(result) {
					requireSuccess(result, _('上传失败'));
					offset = end;
					index++;
					self._checkButton.textContent = _('上传并检查中… %d%%').format(Math.floor(offset * 100 / file.size));
					return sendNext();
				});
			}
			return sendNext().then(function() {
				return api.backup.check({ upload_id: begin.upload_id, filename: file.name, size: file.size });
			});
		}).then(function(result) {
			self._checked = requireSuccess(result, _('备份检查未通过'));
			dom.content(self._checkResult, E('div', { class: 'scn-check-result' }, [
				E('strong', {}, [ _('备份检查通过') ]),
				E('div', { class: 'scn-muted' }, [ _('格式版本：%d，应用版本：%s，文件数：%d').format(result.format_version, result.app_version, result.files) ])
			]));
			notify(_('备份检查通过'));
		}).catch(function(error) {
			dom.content(self._checkResult, []);
			notify(_('检查失败：%s').format(error.message || error), 'danger');
			if (self._uploadId) api.backup.cancelUpload(self._uploadId).catch(function() {});
			self._uploadId = null;
		}).then(function() {
			self._busy = false;
			self._checkButton.textContent = _('检查备份');
			self.updateButtons();
		});
	},

	confirmRestore: function() {
		var self = this;
		ui.showModal(_('确认恢复备份'), [
			E('p', {}, [ _('恢复将覆盖当前配置和数据。系统会先为当前数据创建临时安全备份，恢复完成后可能需要重启服务。') ]),
			E('div', { class: 'right' }, [
				E('button', { class: 'btn', click: ui.hideModal }, [ _('取消') ]), ' ',
				E('button', { class: 'btn cbi-button-negative important', click: function() { ui.hideModal(); return self.restoreBackup(); } }, [ _('确认恢复') ])
			])
		]);
	},

	restoreBackup: function() {
		var self = this;
		this._busy = true;
		this._restoreButton.textContent = _('恢复中…');
		this.updateButtons();
		return api.backup.restore(this._uploadId).then(function(result) {
			requireSuccess(result, _('恢复失败'));
			notify(_('备份恢复成功'));
			self._file = null;
			self._uploadId = null;
			self._checked = null;
			dom.content(self._fileInfo, []);
			dom.content(self._checkResult, []);
		}).catch(function(error) {
			notify(_('恢复失败：%s').format(error.message || error), 'danger');
		}).then(function() {
			self._busy = false;
			self._restoreButton.textContent = _('开始恢复');
			self.updateButtons();
		});
	},

	unload: function() {
		if (this._uploadId) api.backup.cancelUpload(this._uploadId).catch(function() {});
	},

	handleSaveApply: null,
	handleSave: null,
	handleReset: null
});
