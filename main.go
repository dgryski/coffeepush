package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gregdel/pushover"
)

type Event struct {
	Response struct {
		SupporterEmail  string `json:"supporter_email"`
		NumberOfCoffees string `json:"number_of_coffees"`
		TotalAmount     string `json:"total_amount"`
		// SupportCreatedOn string `json:"support_created_on"` // unused
	} `json:"response"`
}

func main() {
	// Create a new pushover app with a token
	apiKey := os.Getenv("PUSHOVER_API_TOKEN")
	if apiKey == "" {
		log.Fatalf("Missing PUSHOVER_API_TOKEN")

	}
	app := pushover.New(apiKey)

	http.HandleFunc("/push/", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.URL.Path, "/push/")

		// TODO(dgryski): authenticate secret token

		decoder := json.NewDecoder(r.Body)
		var event Event
		if err := decoder.Decode(&event); err != nil {
			log.Printf("error decoding json: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		name := event.Response.SupporterEmail
		if name == "" {
			name = "Somebody"
		}
		// TODO(dgryski): coffee vs link
		msgtxt := fmt.Sprintf("%v bought %s coffee(s) for $%s", name, event.Response.NumberOfCoffees, event.Response.TotalAmount)

		recipient := pushover.NewRecipient(token)
		message := pushover.NewMessage(msgtxt)
		_, err := app.SendMessage(message, recipient)
		if err != nil {
			log.Printf("error sending push: %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
			return
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Write(rootTemplateHTML)
	})

	port := ":8080"

	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	log.Println("Listening on port", port)

	log.Fatal(http.ListenAndServe(port, nil))
}

var rootTemplateHTML = []byte(`
<html>
  <head>
  <title>coffeepush</title>
  <style type="text/css">

    @import url(//fonts.googleapis.com/css?family=Droid+Serif);

    body {
       background : lightgrey ;
       margin-top : 100px ;
       font-family : 'Droid Serif' ;
    }

    div#content
    {
       margin : auto ;
       width : 90%;
    }

</style>

  <body>
    <div id="content">
        <h3>buymeacoffee-to-pushover webhook</h3>

This is a webhook for <a href="https://buymeacoffee.com">Buy Me A Coffee's</a> purchase notifications.
It forwards the notification through <a href="http://pushover.net">Pushover</a> to your Android or iOS device.

<p>Sample notification:
<pre>
      user@example.com bought 1 coffee(s) for $3
</pre>

        <p> Setup:

        <ul>
        <li>Install the <a href="https://pushover.net/clients">Pushover app</a> on your phone
        <li>Login to <a href="https://pushover.net/">Pushover</a> and copy your user key from the home page.
        </ul>

	<p> To enable:
        <ul>
        <li>Go to the <a href="https://www.buymeacoffee.com/webhook">Buy Me A Coffee Webhooks</a> page.
        <li>Enter
            <ul><li><b><tt>https://coffeepush.appspot.com/push/YOUR_USER_KEY</tt></b></ul>
        <li>Click <em>Create new webhook</em>
        <li>Click <em>Send Test</em> for instant gratification.
        </ul>

        <p>Bugs and patches: <a href="https://github.com/dgryski/coffeepush">github.com/dgryski/coffeepush</a>

	<p>And of course, if you use this, <a href="https://buymeacoffee.com/dgryski">buy me a coffee</a> :)

    </div>
  </body>
</html>
`)
