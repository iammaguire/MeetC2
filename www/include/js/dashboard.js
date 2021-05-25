var data = [
    {
        'rank': 9,
        'content': 'Alon',
        'UID': '5'
    }, {
        'rank': 9,
        'content': 'Alon',
        'UID': '5'
    },
];

var intervalId = window.setInterval(function () {
    $.ajax({
        type: "GET",
        url: "http://127.0.0.1:8000/api/beacons",
        success: function (d) {
            var data = JSON.parse(d)
            console.log(d)
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
            $('#beacon-table').html(html);        },
        error: function (jqXHR, textStatus, errorThrown) {
            console.log(jqXHR)
        },
    });
}, 500);
