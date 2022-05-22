var beacons = new Array();
var listeners = new Array();
var modules = new Array();
var terminals = new Array();

var selectedModule;
var editor;

window.setInterval(function () {
    $.ajax({
        type: "GET",
        url: "http://" + window.location.host + "/api/beacons",
        success: function (d) {
            jsonToTable(d, "#beacon-table")
        },
        error: function (jqXHR, textStatus, errorThrown) {
            //console.log(jqXHR)
        },
    });

    $.ajax({
        type: "GET",
        url: "http://" + window.location.host + "/api/listeners",
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
        url: "http://" + window.location.host + "/api/updates",
        success: function (d) {
            dec = JSON.parse(d)
            for (var i = 0; i < dec.length; i++) {
                $.notify(dec[i].Title + "\n" + dec[i].Msg, "success");
            }
        },
        error: function (jqXHR, textStatus, errorThrown) {
            //console.log(jqXHR)
        },
    });
}, 1000);

window.setInterval(function () {
    updateModule();
}, 10000);

window.setInterval(function () {
    if (selectedModule != null) {
        selectedModule.source = editor.getValue();
    }
}, 100);

$(document).ready(function() {  
    ace.require("ace/ext/language_tools");
    editor = ace.edit("codeEditor");
    editor.setTheme("ace/theme/dawn");
    editor.session.setMode("ace/mode/csharp");
    editor.setOptions({
        enableLiveAutocompletion: true,
        fontSize: "11pt",
        showPrintMargin: false
    });

    $('#newModuleBtn').click(function() { $("#newModuleModal").modal('show'); });
    $('#newBeaconBtn').click(function() { $("#newBeaconModal").modal('show'); });
    
    $('#validateBtn').click(function() { 
        if (selectedModule != null) { 
            updateModule();
            $.ajax({
                type: "GET",
                url: "http://" + window.location.host + "/api/compile",
                data: { name: selectedModule.Name },
                success: function (d) {
                    if (d == "Good") {
                        $.notify("Compiled without error", "success");
                        $('#codeStackTrace').html("Compiled without error")
                    } else {
                        $.notify("Errors occurred during compilation", "error");
                        $('#codeStackTrace').html(d.replace(/\n/g, "<br />"))
                    }
                },
                error: function (jqXHR, textStatus, errorThrown) {
                    //console.log(jqXHR)
                },
            });
        }
    }) 

    $('#newHTTPListenerBtn').click(function() { 
        $.ajax({
            type: "GET",
            url: "http://" + window.location.host + "/api/netifaces",
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
    });
    
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
        data: {interface: $('#listenerInterface').val().split(' ')[1], hostname: $('#listenerHostname').val(), port: $('#listenerPort').val()},
        success:function(){
            $("#newHTTPListenerModal").modal("hide");
            ShowView(event, 'dashboard')
        }
    });
});

window.onbeforeunload = function(){
    updateModule();
}

window.addEventListener("beforeunload", function(e){
    updateModule();
}, false);

$('#moduleForm').submit(function(e) {
    e.preventDefault();

    var name = $('#moduleName').val();
    var language = $('#moduleLanguage').val();
    var sampleSource = "";

    for (i = 0; i < modules.length; i++) {
        if (modules[i].Name == name) {
            $.notify("Module already exists", "error");
            return
        }
    }

    if (language == "C#") {
        sampleSource = `namespace Module {
    using System;
    
    static class ` + name + ` {
        static void Main(string[] args) {
            Console.WriteLine("Hello world!"); 
        } 
    }\n}`
    } else if (language == "Go") {
        sampleSource = `package main

import "fmt"

func main() {
    fmt.Println("Hello from module!")
}`
    }


    module = { Name: name, Language: language, Source: sampleSource }
    selectedModule = module
    modules.push(module);
           
    editor.setValue(sampleSource, 1);

    updateModule();
    getModules();

    $("#newModuleModal").modal("hide");
    $('#codeEditorContainer').css('display:block');
});

function switchModuleCode(row) {
    updateModule();
    getModules();
    console.log('adasdasdasdasd')
    $('#codeEditorContainer').css('display', 'block');

    moduleName = row.innerHTML.split('<td>')[1].split('</td>')[0];

    for (i = 0; i < modules.length; i++) {
        if (modules[i].Name === moduleName) {
            selectedModule = modules[i];
            editor.setValue(modules[i].Source, 1);

            $('#editorModuleFilename').html(modules[i].Name)
            $('#codeEditorContainer').css('display:block');

            break;
        }
    }
}

function newBeacon() {
    console.log('new beacon!')
}

function updateModule() {
    if (selectedModule != null) {
        selectedModule.Source = editor.getValue();
        $.ajax({
            type: "GET",
            async: false,
            url: "http://" + window.location.host + "/api/updatemodule",
            data: { name: selectedModule.Name, language: selectedModule.Language, source: selectedModule.Source },
        });
    }
}

function getModules() {
    $.ajax({
        type: "GET",
        url: "http://" + window.location.host + "/api/modules",
        success: function (d) {
            modules = new Array();
            data = JSON.parse(d)
            
            for (i = 0; i < data.length; i++) {
                modules.push(data[i])
            }

            filteredData = Array();
            var filter = ({ Name, Source, Language }) => ({ Name, Language });
            
            for(i = 0; i < data.length; i++) {
                filteredData.push(filter(data[i]));
            }

            jsonToTable(JSON.stringify(filteredData), "#module-table", true, "", "switchModuleCode(this)")
        },
        error: function (jqXHR, textStatus, errorThrown) {
            //console.log(jqXHR)
        },
    });
}

function jsonToTable(data, idName, clickable = false, className = "", clickCallback = "") {
    var data = JSON.parse(data)
    var html = '<table class="table table-striped table-hover">';
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
        
        html += '<tr' + (clickCallback != "" ? ' onclick="' + clickCallback + '" ' : '') + (className != '' ? ' class="' + className + '" ' : "") + (clickable ? ' role="button">' : '>');
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
        var socket = new WebSocket("ws://" + window.location.host + "/api/ws");

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