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
    $('#newBeaconBtn').click(function(){ $("#newBeaconModal").modal('show'); }) 
    var beaconElement = $('#mainTerminal')   
    addTerminal(beaconElement, null, "")  
    beaconElement.terminal(function(command) {
        let terminal = this
        addTerminal(this, null, command)  
        try { terminal.socket.send("main::" + command) } catch(error) { console.log(error) }
    }.bind(this), {
        greetings: '',
        name: "main_term",
        height: 700,
        prompt: 'meetC2> ',
    })
    ShowView(null, 'dashboard');
});

function newBeacon() {
    console.log('new beacon!')
}

$('#beaconForm').submit(function(e){
    e.preventDefault();
    $.ajax({
        url: '/api/new',
        type: 'get',
        data: {platform: $('#beaconPlatform').val(), arch: $('#beaconArch').val()},
        success:function(){
            $("#newBeaconModal").modal("hide");
            ShowView(event, 'dashboard')
        }
    });
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

function addTerminal(terminal, beacon, command) {
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
            if (beacon == null) {
                socket.send("main::" + command)
            } else {
                socket.send("beacon:" + beacon.Id + ":" + command)
            }
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
                terminal.echo('\r')
             //   terminal.echo()
//                console.log("'" + event.data + "'")
            }
        }
        terminal.socket = socket
        console.log(terminal.socket, "list")
    }
}

function createBeaconTab(beacon) {
    var bTabDivs = document.getElementById("beaconTabDivs");
    var bTab = document.getElementById("beaconTabId");
    console.log(bTab)
    bTab.innerHTML += '<button class="beaconTabButton" onclick="openBeaconTab(\'' + beacon.Id + '\')">' + beacon.Id + '</button>'
    bTabDivs.innerHTML += '<div id="' + beacon.Id + '" class="beaconTab console" style="display:none"></div>'
    jQuery(function($, undefined) {
        var beaconElement = $('#'+beacon.Id)   
        beaconElement.terminal(function(command) {
            let terminal = this
            addTerminal(this, beacon, command)
            try { terminal.socket.send("beacon:" + beacon.Id + ":" + command) } catch(error) {}
        }, {
            greetings: '',
            name: beacon.Id + "_term",
            height: 700,
            prompt: beacon.Id + '@' + beacon.Ip + '> ',
        })
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