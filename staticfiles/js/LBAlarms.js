//Function initialize_table is used to initialize the table, populate it and deactivate the input buttons and forms
function initialize_alarms(data, mode) {

    $("#myAlarms").dynamicTable({
        columns: [{
            text: "Type of alarm",
            key: "type"
        }, {
            text: "Recipient",
            key: "recipient"
        }, {
            text: "Parameter",
            key: "parameter"
        }, {
            text: "Last active",
            key: "last_active"
        }, {
            text: "Active",
            key: "active"
        }],
        getControl: function (columnKey) {
            var disabled = "disabled='true'";
            if (columnKey == "type") {
                return '<select class="form-control"><option value="minimum">Minimum hosts</option></select>';
            } else if (columnKey == "recipient" || (columnKey == 'parameter')) {
                disabled = ""
            }
            return '<input type="text" class="form-control" ' + disabled + '/>';
        },
        validate: function (key, value) {
            if (key == "recipient") {
                if (!validEmail(value)) {
                    $("#alarms-name-status").html('<img src="/static/js/custom/images/dialog-error.png"</img> The email ' + value + ' is not valid!<img alt="Help" src="/static/js/custom/images/help-browser.png"</img></a><br/>');
                    $('#edit-submit').prop("disabled", true); //Disable submit button
                    return false;
                }
            } else if (key == "parameter") {
                if (!Number.isInteger(parseInt(value))) {
                    $("#alarms-name-status").html($("#alarms-name-status").html() + '<img src="/static/js/custom/images/dialog-error.png"</img> The parameter ' + value + ' is not a number!<img alt="Help" src="/static/js/custom/images/help-browser.png"</img></a> <br/>');
                    $('#edit-submit').prop("disabled", true); //Disable submit button
                    return false;
                }
            }

            return true
        },
        validateSuccess: function () {
            $("#alarms-name-status").html(""); //Hide error message and enable submit button
            $('#edit-submit').prop("disabled", false);
        },

        data: data
    })
    if (mode != "modify") { //If we are not in modify mode then we disable the buttons and fields
        $(".btn").attr("disabled", true);
        $(".form-control").attr("disabled", true);
    }

}
function validEmail(value) {
    if (/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/.test(value)) {
        return true;
    }

    return false;
}


function DisplayAlarms(data) {
    var datalist = [];
    if (data) {
        alarms = data
        for (var i = 0; i < alarms.length; i++) {
            info = (alarms[i]).split(":")
            datalist.push({
                "type": info[0],
                "recipient": info[1],
                "parameter": info[2],
                "active": info[3],
                "last_active": info[4]
            })
        }
    }
    initialize_alarms(datalist, mode);
}

function AlarmsAdded(dataArray, clusterObject) {
    var alarms = [];
    dataArray.forEach(function (entry) {
        alarms.push(entry.type + ":" + entry.recipient + ":" + entry.parameter)
    });
    clusterObject.setAlarms(String(alarms));

    checkSubmit(clusterObject);
}
