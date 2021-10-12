package main

import (
	"bbmachine/handler"
	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/login")
}

func main() {

	h := handler.NewHandler()
	http.HandleFunc("/create_chat", h.CatchErr(h.CreateChat))
	// http.HandleFunc("/join_chat", JoinChat)
	http.HandleFunc("/login", h.CatchErr(h.Login))
	http.HandleFunc("/", home)

	// var addr = flag.String("http://127.0.0.1", ":1718", "http service address")
	fmt.Printf("add %v\n", h.Config.Get("addr"))
	err := http.ListenAndServe(fmt.Sprint(h.Config.Get("addr")), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var info = document.getElementById("info");
    var input = document.getElementById("input");
    var chat = document.getElementById("chat");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}"+ "?user_name="+ info.value);
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RECE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: To: "+chat.value+": " + input.value);
        ws.send(chat.value+"|"+input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<p><input id="info" type="text" value="2">
<button id="open">Open</button>
<button id="close">Close</button>
<p><div>chat room</div><input id="chat" type="text" value="web-dev-sh">
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
