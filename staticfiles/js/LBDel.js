/**
 * @author pmaksymi
 * Updated by Ignacio to use the ermis API getting rid of XML and PHP
 * LBDel - Javascript side code for the DELETE module of LBWeb
 */

(function($){

var newCluster = new LBCluster();
var SelectInitVal = "Please_select";

$(document).ready(function(){
	initialize_form(true, 'delete');

	$('#clusterList').change(function(){
		loadCluster($('#clusterList').val(), newCluster, false);
	});

	$('#edit-submit').click(function(event){
		event.preventDefault();
		submitForm('delete', newCluster);
	});
  checkSubmit(newCluster);

});

})(jQuery)
