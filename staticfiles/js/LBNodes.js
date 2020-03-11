
//Function initialize_table is used to initialize the table, populate it and deactivate the input buttons and forms
function initialize_nodes(data, mode) {

    $("#myTable").dynamicTable({
        columns: [{
            text: "Name",
            key: "name"
        }, {
            text: "Access Control",
            key: "access"
        }],
        getControl: function (columnKey) {
            if (columnKey == "access") {
                return '<select class="form-control"><option value="0">Allow</option><option value="1">Forbidden</option></select>';
            }
            return '<input type="text" class="form-control" />';
        },
        validate: function (key, value) {
            if (key == "name") {
                //Validate input
                if (isNodeRFC952Compliant(value.toLowerCase()) != true) {
                    $("#nodes-name-status").html('<img src="/staticfiles/js/custom/images/dialog-error.png"</img> The node name ' + value + ' is not RFC compliant! <a href="http://configdocs.web.cern.ch/configdocs/dnslb/index.html"><img alt="Help" src="/staticfiles/js/custom/images/help-browser.png"</img></a>');
                    $('#edit-submit').prop("disabled", true); //Disable submit button
                    return false;
                }
            }
            return true;

        },
        validateSucess: function () {
            $("#nodes-name-status").html(""); //Hide error message and enable submit button
            $('#edit-submit').prop("disabled", false);
        },
        data: data
    })
    if (mode != "modify") { //If we are not in modify mode then we disable the buttons and fields
        $(".btn").attr("disabled", true);
        $(".form-control").attr("disabled", true);
    }

}
//FUnction that populates the table with the received data
function DisplayReceivedNodes(AllowedNodes, ForbiddenNodes) {
    var KeyValue = [];
    //Split and filter the allowed nodes, push them in the array
    if (AllowedNodes != null) {
        var allowed = AllowedNodes.split(",").filter(Boolean);
        allowed.forEach(function (entry) {
            KeyValue.push({
                "name": entry,
                "access": 0
            });
        })
    }
    //Split and filter the string of Forbidden nodes, push them on the same array
    if (ForbiddenNodes != null) {
        var forbidden = ForbiddenNodes.split(",").filter(Boolean);
        forbidden.forEach(function (entry) {
            KeyValue.push({
                "name": entry,
                "access": 1
            });
        })
    }
    //Populate the table with the array 
    initialize_nodes(KeyValue, mode)
}


//On submit , gets the data from table, categorizes them and sets them to the cluster object

function HostsAdded(dataArray, clusterObject) {
    var allowed = [];
    var forbidden = [];
    dataArray.forEach(function (entry) {
        if (entry.access === "0") {
            allowed.push(entry.name);
        } else {
            forbidden.push(entry.name);
        }
    });
    clusterObject.setAllowedNodes(String(allowed));
    clusterObject.setForbiddenNodes(String(forbidden));
    checkSubmit(clusterObject);

}
