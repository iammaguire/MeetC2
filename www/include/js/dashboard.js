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
}, 500);

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
        html += '<tr>';
        $.each(value, function (index2, value2) {
            html += '<td>' + value2 + '</td>';
        });
        html += '<tr>';
    });
    html += '</tbody></table>';
    $(idName).html(html);
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