/**
 * @author pmaksymi
 * Updated by Ignacio to use the ermis API getting rid of XML and PHP
 * LBMod - Javascript side code for the MODIFY module of LBWeb
 */

(function($){

var newCluster = new LBCluster();

$(document).ready(function(){
	initialize_form(true, 'modify');

	$('#clusterList').change(function(){
		loadCluster($('#clusterList').val(), newCluster, true);
  });
  $('#nodesList').change(function(){
		loadCluster($('#nodesList').val(), newCluster, true);
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

	$('#edit-submit').click(function(event){
    event.preventDefault();
    //Before submitting , check for dublicates
    var TableData = $("#myTable").getTableData()
    if(hasDuplicate(TableData) == true) {
      $('#edit-submit').prop("disabled",true); //If dublicate, display error and disable submit button
      $("#nodes-name-status").html('<img src="/staticfiles/js/custom/images/dialog-error.png"</img> There is node name duplicate ! <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>');
      return
    } else if(hasDuplicate(TableData) == false){
      HostsAdded(TableData,newCluster); //If no dublicate, proceed with submission
      $('#edit-submit').prop("disabled",false);
      $("#nodes-name-status").html("");
      submitForm('modify', newCluster);
    }
	
	});

});

})(jQuery)

