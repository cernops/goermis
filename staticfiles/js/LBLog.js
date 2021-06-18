/**
 * @author pmaksymi
 * Updated by Ignacio to use the ermis API getting rid of XML and PHP
 * LBDis - Javascript side code for the DISPLAY module of LBWeb
 */

(function ($) {

        var aliasData = [];
        var SelectInitVal = "Please_select";

        $(document).ready(function () {
                loaderWindow();
                getData('../p/api/v1/alias/');

                $('#clusterList').change(function () {
                        location = $('#clusterList').val();
                });

        });

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
                                        .text(element)
                                        .val("https://monit-timber.cern.ch/kibana/app/kibana::/dashboard/9ab64b30-cdcc-11e7-a2b9-5f1755b5a852?_a=(filters:!((meta:(alias:!n,disabled:!f,index:'monit_prod_loadbalancer_*',key:data.cluster,negate:!f,type:phrase,value:" + element + "),query:(match:(data.cluster:(query:" + element + ",type:phrase))))))&_g=(time:(from:now-1h,mode:quick,to:now))"));
                });

                $('#clusterInfo').text("Please select an LB alias to display its log in the monit-timber.cern.ch Kibana server");

        }

})(jQuery)
