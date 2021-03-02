/**
 * @author pmaksymi
 * Generic lbcluster class
 */

//fix for unimplemented indexOf method for IE
if (!Array.indexOf) {
	Array.prototype.indexOf = function (obj) {
		for (var i = 0; i < this.length; i++) {
			if (this[i] == obj) { return i; }
		}
		return -1;
	}
}

function LBCluster() {
	var name = "";
	this.getName = getName;
	this.setName = setName;

	var external = new Boolean(false);
	this.getExternal = getExternal;
	this.setExternal = setExternal;

	var replies = 0;
	this.getReplies = getReplies;
	this.setReplies = setReplies;

	var hostgroup = "";
	this.getHostgroup = getHostgroup;
	this.setHostgroup = setHostgroup;

	var cnames = "";
	this.getCNames = getCNames;
	this.clearCNames = clearCNames;
	this.addCName = addCName;
	this.setCNames = setCNames;

	var tmpcnames = "";
	this.getTmpCNames = getTmpCNames;
	this.setTmpCNames = setTmpCNames;

	var AllowedNodes = "";
	var ForbiddenNodes = "";
	this.getAllowedNodes = getAllowedNodes;
	this.setAllowedNodes = setAllowedNodes;
	this.getForbiddenNodes = getForbiddenNodes;
	this.setForbiddenNodes = setForbiddenNodes;

	this.showCluster = showCluster;
	this.setCluster = setCluster;
	this.clearCluster = clearCluster;

	var alarms = "";
	this.getAlarms = getAlarms;
	this.setAlarms = setAlarms;

	function getName() {
		return name;
	}

	function setName(string) {
		name = String(string);
	}

	function getExternal() {
		return external;
	}

	function setExternal(bool) {
		external = Boolean(bool);
	}

	function getReplies() {
		return replies;
	}

	function setReplies(int) {
		replies = parseInt(int, 10);
	}

	function getHostgroup() {
		return hostgroup;
	}
	function setHostgroup(string) {
		hostgroup = String(string);
	}
	function setCluster(name, visibility, replies, hostgroup, cnames) {
		setName(name);
		setExternal(visibility);
		setReplies(replies);
		setHostgroup(hostgroup);
		setCNames(cnames);
		setTmpCNames(cnames);

	}

	function showCluster() {
		var status = "";
		status = getName() + "<br>";
		status += getExternal() + "<br>";
		//status += getNodes().toString() + "<br>";
		status += getReplies() + "<br>";
		status += getHostgroup() + "<br>";
		status += getCNames().toString() + "<br/>";
		return status;
	}
	function clearCluster() {
		setName('');
		setExternal(false);
		setReplies(0);
		setHostgroup('');
		clearCNames();
		setTmpCNames('');
		setAllowedNodes('');
		setForbiddenNodes('');
	}
	function getCNames() {
		return cnames;
	}
	function setCNames(string) {
		cnames = string.split(",").sort().filter(Boolean).join(",");
	}
	function clearCNames() {
		cnames = "";
	}
	function addCName(alias) {
		/* Let's keep them sorted */
		var tmp = cnames.split(",");
		tmp.push(alias);
		cnames = tmp.sort().filter(Boolean).join(",");
	}
	function getTmpCNames() {
		return tmpcnames;
	}
	function setTmpCNames(string) {
		tmpcnames = string.split(",").sort().filter(Boolean).join(",");
	}
	function getAllowedNodes() {
		return AllowedNodes;
	}
	function getForbiddenNodes() {
		return ForbiddenNodes;
	}
	function setAllowedNodes(string) {
		AllowedNodes = string.split(",").sort().filter(Boolean).join(",");
	}
	function setForbiddenNodes(string) {
		ForbiddenNodes = string.split(",").sort().filter(Boolean).join(",");
	}

	function getAlarms() {
		return alarms;
	}
	function setAlarms(string) {
		alarms = string;
	}

}

