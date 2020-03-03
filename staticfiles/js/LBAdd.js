/**
 * @author pmaksymi
 * Updated by Ignacio to use the ermis API getting rid of XML and PHP
 * LBAdd - Javascript side code for the ADD module of LBWeb
 */

(function($){

 var newCluster = new LBCluster();

 $(document).ready(function(){
  var fastSelect = $('.cnamesInput').fastselect().data('fastselect');

 	initialize_form(false);

	$("#edit-submitted-alias-name").on('input keyup', function(){
		verifyNameChanged($("#edit-submitted-alias-name").val(),newCluster);
	});

	$("#edit-submitted-alias-name").focus(function(){
		$("#edit-submitted-alias-name").keyup();
	});

  $("#edit-submitted-alias-name").on('paste', function(e){
    var pasteData = e.originalEvent.clipboardData.getData('text')
    verifyNameChanged(pasteData, newCluster);
  });

	$('input[name="external"]').click(function(){
		visibilityChanged($('input[name="external"]:checked').val(),newCluster);
	});

	$("#edit-submitted-hostgroup").on('input keyup', function(){
		hostgroupChanged($("#edit-submitted-hostgroup").val(),newCluster);
	});

	$("#edit-submitted-hostgroup").focus(function(){
		$("#edit-submitted-hostgroup").keyup();
	});

  $("#edit-submitted-hostgroup").on('paste', function(e){
    var pasteData = e.originalEvent.clipboardData.getData('text')
    hostgroupChanged(pasteData, newCluster);
  });

	$("#edit-submitted-advanced-best-hosts").keyup(function(){
		bestHostsChanged($("#edit-submitted-advanced-best-hosts").val(),newCluster);
	});

  $("#edit-submitted-advanced-cnames").change(function(){
    cnamesChanged($("#edit-submitted-advanced-cnames").val(),newCluster);
  });

	$('#clearButton').click(function(){
		clearForm(newCluster);
	});

	$("#edit-submit").click(function(event){
		event.preventDefault();
		submitForm('create', newCluster);
	});

 });

$(document).ajaxError(function(event, jqXHR, settings) {
        var contentType = jqXHR.getResponseHeader("Content-Type");
        if ( (contentType === null) && (jqXHR.status === 0) ){
            // assume that our login has expired - reload our current page
            window.location.reload();
        }
 });

function verifyNameChanged(name, clusterObject){
  $("#webform-component-namestatus").html('<img src="/staticfiles/js/spinner.gif"</img> Working..');
  var error="";
  if (name === '')   {
    error = "No name given"
  } else if  (name.length>32){
    error = "The name is longer than 32 characters"
  } else if(isCnameRFC952Compliant(name.toLowerCase())!= true)  {
    error = "Not RFC compliant!"
  }

  if (error != "") {
    $("#webform-component-namestatus")
    .html('<img src="/staticfiles/js/custom/images/dialog-error.png"</img> '+error+' <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
    .css("padding-top","15px").css("padding-left","3px");
    clusterObject.setName('');
    checkSubmit(clusterObject);

  } else {
    $.get('checkname',{hostname: name},function(result){
      evaluateClusterName(name,result,clusterObject);
    });
  }
}


function evaluateClusterName(name, reply, clusterObject){
  var message='<img src="/staticfiles/js/custom/images/dialog-error.png"</img> This name is unavaible ';
	if (reply == 0)	{
    message='<img src="/staticfiles/js/custom/images/dialog-ok.png"</img> '+name+' is avaible';
		clusterObject.setName(name);
	}	else	{
		clusterObject.setName('');
		//Returning success
  }
  $("#webform-component-namestatus")
		.html(message + '<a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>')
		.css("padding-top","15px").css("padding-left","3px");
	checkSubmit(clusterObject);
}
})(jQuery)
