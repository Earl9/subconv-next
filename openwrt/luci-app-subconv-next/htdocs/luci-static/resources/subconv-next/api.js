'use strict';
'require rpc';

var callStatus = rpc.declare({
	object: 'luci.subconv',
	method: 'status',
	expect: { '': {} },
	reject: true
});

var callStart = rpc.declare({
	object: 'luci.subconv',
	method: 'start',
	expect: { '': {} },
	reject: true
});

var callStop = rpc.declare({
	object: 'luci.subconv',
	method: 'stop',
	expect: { '': {} },
	reject: true
});

var callRestart = rpc.declare({
	object: 'luci.subconv',
	method: 'restart',
	expect: { '': {} },
	reject: true
});

var callGetConfig = rpc.declare({
	object: 'luci.subconv',
	method: 'get_config',
	expect: { '': {} },
	reject: true
});

var callSetConfig = rpc.declare({
	object: 'luci.subconv',
	method: 'set_config',
	params: {
		enabled: false,
		listen: '',
		port: 0,
		data_dir: '',
		log_level: '',
		public_base_url: ''
	},
	expect: { '': {} },
	reject: true
});

var callGetAutostart = rpc.declare({
	object: 'luci.subconv',
	method: 'get_autostart',
	expect: { '': {} },
	reject: true
});

var callSetAutostart = rpc.declare({
	object: 'luci.subconv',
	method: 'set_autostart',
	params: { enabled: false },
	expect: { '': {} },
	reject: true
});

var callStorageCheck = rpc.declare({
	object: 'luci.subconv',
	method: 'storage_check',
	expect: { '': {} },
	reject: true
});

var callLogsRead = rpc.declare({
	object: 'luci.subconv',
	method: 'logs_read',
	params: { source: '', lines: 0 },
	expect: { '': {} },
	reject: true
});

var callLogsClear = rpc.declare({
	object: 'luci.subconv',
	method: 'logs_clear',
	expect: { '': {} },
	reject: true
});

var callLogsDownload = rpc.declare({
	object: 'luci.subconv',
	method: 'logs_download',
	params: { source: '' },
	expect: { '': {} },
	reject: true
});

var callBackupExport = rpc.declare({
	object: 'luci.subconv',
	method: 'backup_export',
	expect: { '': {} },
	reject: true
});

var callBackupExportCleanup = rpc.declare({
	object: 'luci.subconv',
	method: 'backup_export_cleanup',
	params: { id: '' },
	expect: { '': {} },
	reject: true
});

var callBackupUploadBegin = rpc.declare({
	object: 'luci.subconv',
	method: 'backup_upload_begin',
	expect: { '': {} },
	reject: true
});

var callBackupUploadChunk = rpc.declare({
	object: 'luci.subconv',
	method: 'backup_upload_chunk',
	params: { upload_id: '', index: 0, data: '' },
	expect: { '': {} },
	reject: true
});

var callBackupCheck = rpc.declare({
	object: 'luci.subconv',
	method: 'backup_check',
	params: { upload_id: '', filename: '', size: 0 },
	expect: { '': {} },
	reject: true
});

var callBackupRestore = rpc.declare({
	object: 'luci.subconv',
	method: 'backup_restore',
	params: { upload_id: '' },
	expect: { '': {} },
	reject: true
});

var callBackupUploadCancel = rpc.declare({
	object: 'luci.subconv',
	method: 'backup_upload_cancel',
	params: { upload_id: '' },
	expect: { '': {} },
	reject: true
});

function settle(promise) {
	return promise.then(function(value) {
		return { ok: value != null, value: value };
	}).catch(function(error) {
		return { ok: false, error: error };
	});
}

return L.Class.extend({
	service: {
		status: callStatus,
		start: callStart,
		stop: callStop,
		restart: callRestart
	},
	config: {
		get: callGetConfig,
		save: function(values) {
			return callSetConfig(values || {});
		}
	},
	system: {
		getAutostart: callGetAutostart,
		setAutostart: function(enabled) {
			return callSetAutostart({ enabled: enabled });
		}
	},
	storage: {
		check: callStorageCheck
	},
	logs: {
		read: function(source, lines) {
			return callLogsRead({ source: source, lines: lines });
		},
		clear: callLogsClear,
		download: function(source) {
			return callLogsDownload({ source: source });
		}
	},
	backup: {
		exportArchive: callBackupExport,
		cleanupExport: function(id) {
			return callBackupExportCleanup({ id: id });
		},
		beginUpload: callBackupUploadBegin,
		uploadChunk: function(uploadId, index, data) {
			return callBackupUploadChunk({ upload_id: uploadId, index: index, data: data });
		},
		check: function(values) {
			return callBackupCheck(values || {});
		},
		restore: function(uploadId) {
			return callBackupRestore({ upload_id: uploadId });
		},
		cancelUpload: function(uploadId) {
			return callBackupUploadCancel({ upload_id: uploadId });
		}
	},
	settle: settle
});
