/* Common functions used by Add/Display/Modify/Delete */

function loaderWindow(action) {
  if (action == 'close') {
    loaderdialog.dialog('close'); return;
  } else {
    loaderdialog = $('<div></div>')
      .html("<div id='loadingscreen'>Please wait..</div>")
      .css("padding-left", "25px").css("background", "/staticfiles/js/spinner.gif no-repeat 5px 8px")
      .dialog({
        autoOpen: false,
        title: 'Loading',
        modal: true,
        resizable: false,
        buttons: {},
        closeOnEscape: false,
        width: 460,
        minHeight: 50,
        draggable: false
      });
    loaderdialog.dialog('open');
  }
}
var aliasData = [];
var SelectInitVal = "Please_select";
var fastSelect = null;
var preventCNamesEdit = true;
var currentCNames = [];
var mode = null; //Global variable to keep track of the action that is being performed (modify/del/display).It is used to manage the table 

/*  Prepares the form. Parameters:
 *  retrieveData (boolean): If true, it will do a query to retrieve the list of all aliases
 *
 *  */
function initialize_form(retrieveData, action) {
  loaderWindow();
  $('.cnamesInput').attr("disabled", true);
  $('.cnamesInput').attr("placeholder", " ");
  mode = action; //keep up to date the global variable

  fastSelect = $('.cnamesInput').fastselect().data('fastselect');
  //reset the form, some browser have a nasty habit of leaving stuff behind
  //clearForm();
  if (retrieveData) {
    getData('https://goermis.cern.ch/lbweb/api/v1/aliases');

  } else {
    loaderWindow('close');
  }

  $('input[name="external"]').click();

  //Initialize table
  initialize_nodes("", mode)


  //make the user aware that he has to select a cluster to edit before entering any details
  toggleEditing(!retrieveData, action);
  checkSubmit();
}
function getClusterNames(aliasData) {

  var clusterNames = [];

  jQuery.each(aliasData, function (index, element) {
    clusterNames.push(element.alias_name.toLowerCase());
  });

  clusterNames.sort();
  clusterNames.unshift(SelectInitVal);
  return clusterNames;
}

function loadClusterList(array) {
  var clusterNames = array;
  jQuery.each(clusterNames, function (index, element) {
    $("#clusterList")
      .append($("<option></option>")
        .attr(element, index)
        .text(element));
  });
}

//Grabs the data from the API
function getData(URI) {
  $.get(URI, { format: 'json', limit: 0 }, function (result) {
    jQuery.each(result.objects, function (index, element) {
      aliasData.push(element);
    });
    ClusterNameList = getClusterNames(aliasData);
    loadClusterList(ClusterNameList);
    loaderWindow('close');
  });
}

/*  What happens when an alias is selected
*/
function loadCluster(name, newCluster, editable) {
  if (name === SelectInitVal) {
    clearForm(newCluster);
    $('#edit-submit').prop("disabled", true);

  } else {
    newCluster.clearCluster();
    populateClass(name, newCluster);
    currentCNames = newCluster.getCNames().split(',');
    toggleEditing(editable);
    writeFields(newCluster);
    checkSubmit(newCluster);
  }
}

/* Returns the data for a particular alias*/
function getClusterAliasData(alias_name) {
  for (var i = 0; i <= aliasData.length + 1; i++) {
    if (aliasData[i].alias_name == alias_name) {
      return aliasData[i];
    }
  }
  return;
}
/*COMMON METHODS FOR CHANGES IN FIELDS*/
//Check for duplicate nodes
var hasDuplicate = function (array) {

  return array.map(function (value) {
    return value.name

  }).some(function (value, index, array) {
    return array.indexOf(value) !== array.lastIndexOf(value);
  })
}

function visibilityChanged(string, clusterObject) {
  var external = false;
  var message = 'only inside the CERN network';

  if (string === "yes") {
    message = 'on the CERN internal network and on the internet'
    external = true;
  }
  $("#webform-component-visibilitystatus").html('This cluster will be avaible \
    ' + message + ' \
    <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
    ;
  clusterObject.setExternal(external);
}

function hostgroupChanged(string, clusterObject) {
  var hostgroup = string;

  $("#webform-component-hostgroupstatus")
    .html('<img src="/staticfiles/js/spinner.gif"</img> Working..')


  if (hostgroup === '') {
    $("#webform-component-hostgroupstatus")
      .html('<img src="/staticfiles/js/custom/images/dialog-warning.png"</img> No hostgroup given <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
      .css("padding-top", "15px").css("padding-left", "3px");
    clusterObject.setHostgroup('');
    checkSubmit(clusterObject);
    return;
  } else {
    var HGRegex = new RegExp("^[a-z][a-z0-9\_\/]*[a-z0-9]$");
    if (HGRegex.test(hostgroup)) {

      $("#webform-component-hostgroupstatus")
        .html('<img src="/staticfiles/js/custom/images/dialog-ok.png"</img> taking ' + hostgroup + ' as the hostgroup. You have to be administrator of ' + hostgroup + '. Alias members will have to belong to the base hostgroup. <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
        .css("padding-top", "15px").css("padding-left", "3px");
      clusterObject.setHostgroup(hostgroup);
      checkSubmit(clusterObject);
      return;
    } else {
      $("#webform-component-hostgroupstatus")
        .html('<img src="/staticfiles/js/custom/images/dialog-error.png"</img> Not a valid hostgroup <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
        .css("padding-top", "15px").css("padding-left", "3px");
      clusterObject.setHostgroup('');
      checkSubmit(clusterObject);
      return;
    }
  }
}
//Validation function
function isCnameRFC952Compliant(Cname) {
  if (Cname.length < 2 || Cname.length > 511) { return false; }
  /* Allowing also subdomains */
  var RFCRegex = new RegExp(/^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/); // RFC 1123
  return RFCRegex.test(Cname);
}

function isNodeRFC952Compliant(Node) {
  var check = true;
  var segmented = Node.split(".");
  if (segmented.length < 3 || segmented.length > 10) {
    check = false;
  }
  var RFCRegex = new RegExp("^[a-z][a-z0-9\-]*[a-z0-9]$");
  segmented.forEach(function (segment) {
    if (segment.length < 2 || segment.length > 32) {
      check = false;
    }
    else if (!RFCRegex.test(segment)) {
      check = false;
    }
  });
  return check;
}

function isNodeRFC952Compliant(Node) {
  var check = true;
  var segmented = Node.split(".");
  if (segmented.length < 3 || segmented.length > 10) {
    check = false;
  }
  var RFCRegex = new RegExp("^[a-z][a-z0-9\-]*[a-z0-9]$");
  segmented.forEach(function (segment) {
    if (segment.length < 2 || segment.length > 32) {
      check = false;
    }
    else if (!RFCRegex.test(segment)) {
      check = false;
    }
  });
  return check;
}

/* And these are also common */

function toggleEditing(bool, action) {
  if (bool === false) {
    $('#edit-submitted-alias-name').attr("disabled", true);
    $('input[name="external"]').attr("disabled", true);
    $("#edit-submitted-hostgroup").attr("disabled", true);
    $('#edit-submitted-advanced-best-hosts').attr("disabled", true);
    //$('#clusterList').css("border","red","2px","solid")
    $('#clusterInfo').text("Please select an LB alias to " + action);
    $('#edit-submitted-advanced-cnames').attr("disabled", true);
    $('.fstQueryInput')[0].disabled = true;

  }
  else {
    //$('#edit-submitted-alias-name').attr("disabled", false);
    $('input[name="external"]').attr("disabled", false);
    $("#edit-submitted-hostgroup").attr("disabled", false);
    $('#edit-submitted-advanced-best-hosts').attr("disabled", false);
    //$('#clusterList').css("border","red","2px","solid")
    $('#clusterInfo').text("");
    $('#edit-submitted-advanced-cnames').removeAttr("disabled");
    $('.fstQueryInput')[0].disabled = false;
    preventCNamesEdit = false;
  }

}

function populateClass(name, clusterObject) {
  var cluster = getClusterAliasData(name);

  //get the values
  var name = cluster.alias_name;
  var visibility = cluster.external;
  var replies = cluster.best_hosts;
  var hostgroup = cluster.hostgroup;
  var cnames = cluster.cnames;


  if (visibility == "yes") { visibility = true; }
  else { visibility = false; }
  //alert(JSON.stringify(cluster));
  clusterObject.setCluster(name, visibility, replies, hostgroup, cnames);
  DisplayReceivedNodes(cluster.AllowedNodes, cluster.ForbiddenNodes);

  return;
}

function writeFields(clusterObject) {
  //Name
  $('#edit-submitted-alias-name').val(clusterObject.getName());
  nameChanged(clusterObject.getName(), clusterObject);

  //Visibility
  if (clusterObject.getExternal() === true) { $('input[name="external"]')[0].checked = true; }
  else { $('input[name="external"]')[1].checked = true; }

  visibilityChanged($('input[name="external"]:checked').val(), clusterObject);

  //Hostgroup
  $('#edit-submitted-hostgroup').val(clusterObject.getHostgroup());
  hostgroupChanged(clusterObject.getHostgroup(), clusterObject);

  //Replies
  $('#edit-submitted-advanced-best-hosts').val(clusterObject.getReplies());

  //Log
  if (clusterObject.getName() === '') {
    $('#alias-log-headtext').text("");
    $('#alias-log').prop("href", "");
    $('#alias-log').text("");
  } else {
    $('#alias-log-headtext').html("<strong>LB Alias Log</strong>");
    var loglink = "https://monit-timber.cern.ch/kibana/app/kibana::/dashboard/9ab64b30-cdcc-11e7-a2b9-5f1755b5a852?_a=(filters:!((meta:(alias:!n,disabled:!f,index:'monit_prod_loadbalancer_*',key:data.cluster,negate:!f,type:phrase,value:" + clusterObject.getName() + "),query:(match:(data.cluster:(query:" + clusterObject.getName() + ",type:phrase))))))&_g=(time:(from:now-1h,mode:quick,to:now))";

    $('#alias-log').prop("href", loglink);
    $('#alias-log').text(loglink);
  }

  $('#edit-submitted-advanced-cnames').val(clusterObject.getCNames());
  if (fastSelect != null) {
    $('.fstChoiceItem').remove();
    fastSelect.optionsCollection.selectedValues = {};
    clusterObject.getCNames().split(',').forEach(function (element) {
      if (element) {
        fastSelect.addChoiceItem({ 'text': element, 'value': element });
        fastSelect.optionsCollection.setSelected({ 'text': element, 'value': element });
        //        $('.fstChoiceRemove')[0].disabled=true;
        $('.fstChoiceRemove').attr('disabled', preventCNamesEdit);
      }
    });
  }
  $('#edit-submitted-advanced-cnames').attr('initialValue', clusterObject.getCNames());
  cnamesChanged(clusterObject.getCNames(), clusterObject);
  //Write on the hidden forms the node names, prepare for submission
  $("#AllowedNodes").val(clusterObject.getAllowedNodes());
  $("#ForbiddenNodes").val(clusterObject.getForbiddenNodes());


}
function nameChanged(name, clusterObject) {

  $("#NameStatus").html('<img src="/staticfiles/js/spinner.gif"</img> Working..')

  if (name === '') {
    $("#NameStatus")
      .html('<img src="/staticfiles/js/custom/images/dialog-warning.png"</img> No name given <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
      .css("padding-top", "15px").css("padding-left", "3px");
    clusterObject.setName('');
    checkSubmit();
  } else {
    $('#NameStatus')
      .html('<img src="/staticfiles/js/custom/images/dialog-ok.png"</img> Primary key <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
      .css("padding-top", "15px").css("padding-left", "3px");
  }

}

function bestHostsChanged(int, clusterObject) {
  var replies = parseInt(int, 10);
  var msg = '<img src="/staticfiles/js/custom/images/dialog-error.png"</img> ' + replies + ' is not OK';

  //if (replies <= clusterObject.getNodes().length && replies >= 0) {
  if (replies > 0 || replies === -1) {
    msg = '<img src="/staticfiles/js/custom/images/dialog-ok.png"</img> ' + replies + ' is OK';
    clusterObject.setReplies(replies);
    $("#edit-submitted-advanced-best-hosts").val(clusterObject.getReplies());
  } else {
    clusterObject.setReplies(0);
  }
  $("#webform-component-besthoststatus")
    .html(msg + ' <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
    .css("padding-top", "15px").css("padding-left", "3px");
  checkSubmit(clusterObject);
}


function cnamesChanged(cnames_string, clusterObject) {
  var cnames = cnames_string.split(',');

  /* Let's empty the list of possible cnames, since they have to be checked*/
  clusterObject.setTmpCNames(cnames_string);
  clusterObject.clearCNames();

  $("#webform-component-cnamesstatus").html("");
  if (cnames == "") {
    return;
  }
  var waitForCheck = false;
  $("#webform-component-cnamesstatus").html("");

  cnames.forEach(function (my_cname) {
    if (isCnameRFC952Compliant(my_cname.toLowerCase()) != true) {
      $("#webform-component-cnamesstatus")
        .html('<img src="/staticfiles/js/custom/images/dialog-error.png"</img> The cname ' + my_cname + ' is not RFC compliant! <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>');
      return
    }
    if (currentCNames.indexOf(my_cname) >= 0) {
      var current_message = $("#webform-component-cnamesstatus").html();
      $("#webform-component-cnamesstatus")
        .html(current_message + '<br/><img src="/staticfiles/js/custom/images/dialog-ok.png"</img> The cname ' + my_cname + ' already points to the alias <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>');
      clusterObject.addCName(my_cname);

    } else {
      $.get('checkname', { hostname: my_cname }, function (result) {
        evaluateCName(my_cname, result, clusterObject);
      });
      waitForCheck = true;
    }
  });
  if (!waitForCheck) {
    checkSubmit(clusterObject);
  }

}

function evaluateCName(my_cname2, reply, clusterObject) {
  var current_message = $("#webform-component-cnamesstatus").html();
  var new_message = '<br/><img src="/staticfiles/js/custom/images/dialog-error.png"</img>  The name ' + my_cname2 + " is not available";
  if (reply == 0) {
    new_message = '<br/><img src="/staticfiles/js/custom/images/dialog-ok.png"</img> ' + my_cname2 + ' is available';
    clusterObject.addCName(my_cname2);
  }
  $("#webform-component-cnamesstatus")
    .html(current_message + new_message).css("padding-top", "15px").css("padding-left", "3px");

  checkSubmit(clusterObject);
}


function checkSubmit(clusterObject) {
  var submitEnabled = true;
  if (clusterObject == null) {
    submitEnabled = false;
  } else {
    if (clusterObject.getName().length == 0) { submitEnabled = false }
    if (clusterObject.getHostgroup().length == 0) { submitEnabled = false }
    if (clusterObject.getReplies() < -1 || clusterObject.getReplies() == 0) { submitEnabled = false }
    if (clusterObject.getTmpCNames() != clusterObject.getCNames()) { submitEnabled = false }


  }


  $('#edit-submit').prop("disabled", !submitEnabled);
}

function clearForm(clusterObject) {
  clusterObject.clearCluster();
  writeFields(clusterObject);
  initialize_nodes("", mode)
  //fix up what was left behind

}


async function AsyncSubmit(newCluster) {
  writeFields(newCluster);
  await new Promise(resolve => setTimeout(resolve, 500));
  $("#webform-client-form-3").submit();
  $submitDialog.dialog("close");
}


function submitForm(action, newCluster) {

  var textHTML = "You are about to " + action + ":<br>" +
    "Name: " + newCluster.getName() + "<br>" +
    "External: " + newCluster.getExternal() + "<br>" +
    "Hostgroup: " + newCluster.getHostgroup() + "<br>" +
    "Best Hosts: " + newCluster.getReplies() + "<br>"


  if (newCluster.getCNames() != "") {
    textHTML += "Canonical names: " + newCluster.getCNames() + "<br/>";
  }
  if (newCluster.getAllowedNodes() != "") {
    textHTML += "Allowed Hosts:" + newCluster.getAllowedNodes() + "<br/>";
  }
  if (newCluster.getForbiddenNodes() != "") {
    textHTML += "Forbidden Hosts:" + newCluster.getForbiddenNodes() + "<br/>";
  }


  var $submitDialog = $('<div></div>')
    .html(textHTML)
    .dialog({
      autoOpen: false,
      title: 'Submit',
      modal: true,
      resizable: true,
      buttons: {
        "Cancel": function () { $(this).dialog("close"); },
        "Send": function () { AsyncSubmit(newCluster); }
      }
    });

  $submitDialog.dialog('open');
}


