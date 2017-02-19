Webhooks
========

Webhooks allow squIRCy2 to respond to data received by HTTP endpoints.

Webhooks are validated by hashing and signing the payload contents with SHA1 and the generated Webhook key. squIRCy2 will look for this  signature in the Webhook's configured Signature header.

Below is an example Go implementation that hashes and signs the payload before submitting it.

```go
package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"io/ioutil"
)

func main() {
	// The Webhook's ID
	id := "8519993742264042640"
 	// The Webhook's generated key
	key := "c6983a9f-fbd1-4b9c-67fb-5e4e48a7a838"
	// The Webhook's signature header
	header := "X-Signature"
	// The desired payload
	payload := "Hello, World!"
	
	// Hash and sign the payload
	mac := hmac.New(sha1.New, []byte(key))
	_, err := mac.Write([]byte(payload))
	if err != nil {
		panic(err)
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	// Create the web request
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:3000/webhooks/%s", id), strings.NewReader(payload))
	if err != nil {
		panic(err)
	}
	
	// Set the signature header to the generated signature
	req.Header.Add(header, fmt.Sprintf("sha1=%s", signature))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	// Outputs: OK
}
```

A simple handler for the above configuration might look like:

```js
bind('hook.8519993742264042640', function(e) {
    console.log(e.Body);
    // Outputs: Hello, World!
});
```