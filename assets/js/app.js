var webSocket = new WebSocket('ws://localhost:3000/echo');

webSocket.onerror = function(event) {
    onError(event)
};

$('#add-answer').bind('click', function () {
    $('#answers').append('<input type="text" class="form-control" placeholder="Ответ" name="answers[]">');
});

$('#remove-answer').bind('click', function () {
    $('#answers').find('input:last').remove();
});

$("#create-answer").submit(function( event ) {
    var formData = $(this).serializeJSON();
    var jsonData = JSON.stringify(formData);

    webSocket.send(jsonData);

    webSocket.onmessage = function(evt) {
    		if (evt.data == "OK") {
    			$("#createItemModal").modal("hide");
    			$("input").val('');
    		}
    		console.log(evt.data);
    		var arr = JSON.parse(evt.data);

    		for (var i = 0; i < arr.length; i++) {
    		    var ans = arr[i].Answer;
                var qst = arr[i].Question;
                $("#mainTable tbody").append('<tr><th scope="row">'+ i +'</th><td>'+ qst +'</td><td>'+ ans +'</td><td>Не активен</td><td></td></tr>');
            }



        }

    event.preventDefault();
});

