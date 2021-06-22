var beacons = new Array();
var terminals = new Array();

var intervalId = window.setInterval(function () {
    $.ajax({
        type: "GET",
        url: "http://127.0.0.1:8000/api/beacons",
        success: function (d) {
            jsonToTable(d, "#beacon-table")
        },
        error: function (jqXHR, textStatus, errorThrown) {
            console.log(jqXHR)
        },
    });
    $.ajax({
        type: "GET",
        url: "http://127.0.0.1:8000/api/listeners",
        success: function (d) {
            jsonToTable(d, "#listener-table")
        },
        error: function (jqXHR, textStatus, errorThrown) {
            console.log(jqXHR)
        },
    });
    $.ajax({
        type: "GET",
        url: "http://127.0.0.1:8000/api/updates",
        success: function (d) {
            dec = JSON.parse(d)
            for (var i = 0; i < dec.length; i++) {
                $.notify(dec[i].Title + "\n" + dec[i].Msg, "info");
            }
        },
        error: function (jqXHR, textStatus, errorThrown) {
            console.log(jqXHR)
        },
    });
}, 1000);

$(document).ready(function() {
    ShowView(null, 'dashboard');
});

function jsonToTable(data, idName) {
    var data = JSON.parse(data)
    var html = '<table class="table table-striped">';
    html += '<thead><tr>';
    var flag = 0;
    $.each(data[0], function (index, value) {
        html += '<th>' + index + '</th>';
    });
    html += '</tr></thead><tbody>';
    $.each(data, function (index, value) {
        if (value['Ip']) { // Beacon data
            var found = false
            for (i = 0; i < beacons.length; i++) {
                if (beacons[i].Id === value.Id) {
                    found = true
                    break
                }
            }
            if (!found) {
                createBeaconTab(value)
                beacons.push(value)
            }
        }
        
        html += '<tr>';
        $.each(value, function (index2, value2) {
            html += '<td>' + value2 + '</td>';
        });
        html += '<tr>';
    });
    html += '</tbody></table>';
    $(idName).html(html);
}

function openBeaconTab(beaconId) {
    var i;
    var x = document.getElementsByClassName("beaconTab");
    for (i = 0; i < x.length; i++) {
      x[i].style.display = "none";
    }
    document.getElementById(beaconId).style.display = "block";
}

function addTerminal(terminal, beacon) {
    var found = false;
    for (i = 0; i < terminals.length; i++) {
        if (terminal === terminals[i]) {
            found = true;
            break;
        }
    }
    if (!found) {
        terminals.push(terminal);
        var socket = new WebSocket("ws://127.0.0.1:8000/api/ws");
        
        socket.onopen = () => {
            console.log("Successfully Connected");
            socket.send("meetc2:" + beacon.Id)
        };
        
        socket.onclose = event => {
            console.log("Socket Closed Connection: ", event);
            socket.send("Client Closed!")
        };

        socket.onerror = error => {
            console.log("Socket Error: ", error);
        };

        socket.onmessage = event => {
            if (event.data.length > 0) {
                terminal.echo(event.data)
             //   terminal.echo()
//                console.log("'" + event.data + "'")
            }
        }

        terminal.socket = socket
    }
}

function createBeaconTab(beacon) {
    var bTabDivs = document.getElementById("beaconTabDivs");
    var bTab = document.getElementById("beaconTabId");
    console.log(bTab)
    bTab.innerHTML += '<button class="beaconTabButton" onclick="openBeaconTab(\'' + beacon.Id + '\')">' + beacon.Id + '</button>'
    bTabDivs.innerHTML += '<div id="' + beacon.Id + '" class="beaconTab console" style="display:none"></div>'
    jQuery(function($, undefined) {
        $('#'+beacon.Id).terminal(function(command) {
            let terminal = this
            addTerminal(terminal, beacon)
            try { terminal.socket.send("beacon:" + beacon.Id + ":" + command) } catch(error) {}
        }, {
            greetings: '',
            name: beacon.Id + "_term",
            height: 600,
            prompt: beacon.Id + '@' + beacon.Ip + '> '
        });
    });
}

function ShowView(evt, tabname) {
    var i, tabcontent, tablinks;

    tabcontent = document.getElementsByClassName("tabcontent");
    for (i = 0; i < tabcontent.length; i++) {
        tabcontent[i].style.display = "none";
    }

    tablinks = document.getElementsByClassName("tablinks");
    for (i = 0; i < tablinks.length; i++) {
        tablinks[i].className = tablinks[i].className.replace(" active", "");
    }

    document.getElementById(tabname).style.display = "block";
    evt.currentTarget.className += " active";
}