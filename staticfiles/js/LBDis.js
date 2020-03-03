/**
 * @author pmaksymi
 * Updated by Ignacio to use the ermis API getting rid of XML and PHP
 * LBDis - Javascript side code for the DISPLAY module of LBWeb
 */

(function($){

var newCluster = new LBCluster();

$(document).ready(function(){
	initialize_form(true, 'display');


	$('#clusterList').change(function(){
		loadCluster($('#clusterList').val(), newCluster, false);
	});
});


})(jQuery)
