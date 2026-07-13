'use strict';

function isIPv4(value) {
	var parts;

	if (!/^\d+\.\d+\.\d+\.\d+$/.test(value))
		return false;

	parts = value.split('.');
	for (var i = 0; i < parts.length; i++) {
		if (parts[i].length > 3 || Number(parts[i]) > 255)
			return false;
	}

	return true;
}

function isIPv6(value) {
	if (value.indexOf(':') === -1 || !/^[0-9A-Fa-f:.]+$/.test(value))
		return false;

	try {
		return new URL('http://[' + value + ']/').hostname !== '';
	}
	catch (error) {
		return false;
	}
}

function isHostname(value) {
	var normalized = value.charAt(value.length - 1) === '.' ? value.slice(0, -1) : value;
	var labels;

	if (!normalized || normalized.length > 253 || /^\d+(?:\.\d+)*$/.test(normalized))
		return false;

	labels = normalized.split('.');
	for (var i = 0; i < labels.length; i++) {
		if (!/^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$/.test(labels[i]))
			return false;
	}

	return true;
}

function isListenAddress(value) {
	if (typeof value !== 'string' || value === '' || value !== value.trim())
		return false;

	return isIPv4(value) || isIPv6(value) || isHostname(value);
}

function isPort(value) {
	var port;

	if (!/^\d+$/.test(String(value)))
		return false;

	port = Number(value);
	return port >= 1 && port <= 65535;
}

function isPublicBaseURL(value) {
	var url;

	if (value === '')
		return true;
	if (typeof value !== 'string' || value !== value.trim())
		return false;

	try {
		url = new URL(value);
		return url.protocol === 'http:' || url.protocol === 'https:';
	}
	catch (error) {
		return false;
	}
}

return L.Class.extend({
	isIPv4: isIPv4,
	isIPv6: isIPv6,
	isHostname: isHostname,
	isListenAddress: isListenAddress,
	isPort: isPort,
	isPublicBaseURL: isPublicBaseURL
});
