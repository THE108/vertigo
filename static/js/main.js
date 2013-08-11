(function($) {
    "use strict"

    $(document).ready(function() {
        console.log("Start");

        var log = $("#log"),
            msg = $("#message");

        function addMessage(message) {
            log.append($("<div/>", {"class": "alert alert-success", "html": message}));
        }

        var sock = new WebSocket("ws://localhost:8000/ws");
        //sock.binaryType = 'blob'; // can set it to 'blob' or 'arraybuffer
        console.log("Websocket - status: " + sock.readyState);
        sock.onopen = function(m) {
            console.log("CONNECTION opened..." + this.readyState);
        }
        sock.onmessage = function(m) {
            addMessage(m.data)
        }
        sock.onerror = function(m) {
            console.log("Error occured sending..." + m.data);
        }
        sock.onclose = function(m) {
            console.log("Disconnected - status " + this.readyState);
        }

        $("#send-btn").click(function() {
            var ms = JSON.stringify({
                "Port": +msg.val(),
                "Name": "Test"
            });
            sock.send(ms);
        });
    });
})(jQuery);