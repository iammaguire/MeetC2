var beacons = new Array();
var listeners = new Array();
var terminals = new Array();

var intervalId = window.setInterval(function () {
    $.ajax({
        type: "GET",
        url: "http://127.0.0.1:8000/api/beacons",
        success: function (d) {
            jsonToTable(d, "#beacon-table")
        },
        error: function (jqXHR, textStatus, errorThrown) {
            //console.log(jqXHR)
        },
    });
    $.ajax({
        type: "GET",
        url: "http://127.0.0.1:8000/api/listeners",
        success: function (d) {
            listeners = new Array();
            data = JSON.parse(d)
            for (i = 0; i < data.length; i++) {
                listeners.push(data[i])
            }
            jsonToTable(d, "#listener-table")
        },
        error: function (jqXHR, textStatus, errorThrown) {
            //console.log(jqXHR)
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
            //console.log(jqXHR)
        },
    });
}, 1000);

$(document).ready(function() {    
    $('#newBeaconBtn').click(function() { $("#newBeaconModal").modal('show'); }) 
    $('#newHTTPListenerBtn').click(function() { 
        $.ajax({
            type: "GET",
            url: "http://127.0.0.1:8000/api/netifaces",
            success: function (d) {
                data = d.split(/\n/)
                $('#listenerInterface').empty()
                for (i = 0; i < data.length; i++) {
                    if (data[i].length > 0) {
                        $('#listenerInterface').append("<option>" + data[i] + "</data>")
                    }
                }
            },
            error: function (jqXHR, textStatus, errorThrown) {
                console.log(jqXHR)
            },
        });

        $("#newHTTPListenerModal").modal('show'); 
    }) 
    
    var beaconElement = $('#mainTerminal')   
    beaconElement.terminal(function(command) {
        let terminal = this
        addTerminal(terminal, null, command)  
        try { terminal.socket.send("main::" + command) } catch(error) { console.log(error) }
    }.bind(this), {
        greetings: '',
        name: "main_term",
        height: 700,
        prompt: 'meetC2> ',
    })
    addTerminal(beaconElement, null, "")  
    ShowView(null, 'dashboard');
});

function newBeacon() {
    console.log('new beacon!')
}

$('#beaconForm').submit(function(e){
    e.preventDefault();
    $.ajax({
        url: '/api/newbeacon',
        type: 'get',
        data: {platform: $('#beaconPlatform').val(), arch: $('#beaconArch').val()},
        success:function(){
            $("#newBeaconModal").modal("hide");
            ShowView(event, 'dashboard')
        }
    });
});

$('#httpListenerForm').submit(function(e) {
    e.preventDefault();

    if (!/^-?\d+$/.test($('#listenerPort').val())) {
        $.notify("Enter number for port")
        return
    }

    if (parseInt($('#listenerPort').val()) < 1 || parseInt($('#listenerPort').val()) > 65535) {
        $.notify("Port not in range 1-65535", "error")
        return
    }

    for (i = 0; i < listeners.length; i++) {
        if (listeners[i].Port == $('#listenerPort').val()) {
            $.notify("Port already in use", "error");
            return
        }
    }

    $.ajax({
        url: '/api/newhttplistener',
        type: 'get',
        data: {interface: $('#listenerInterface').val().split(' ')[0], hostname: $('#listenerHostname').val(), port: $('#listenerPort').val()},
        success:function(){
            $("#newHTTPListenerModal").modal("hide");
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
                if (terminal.echo != null) {
                    terminal.echo(event.data)
                    terminal.echo('')
                }
            }
        }
        terminal.socket = socket
    }
}

function createBeaconTab(beacon) {
    var bTabDivs = document.getElementById("beaconTabDivs");
    var bTab = document.getElementById("beaconTabId");
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
    
    if (evt != null) {
        evt.currentTarget.className += " active";
    }
}