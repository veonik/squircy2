Configuring
===========

Once the bot is up and running, you can access the web management interface via `localhost:3000`. From this interface you can write scripts and use a web-based REPL, as well as configure squIRCy2 to your liking.

### Configuration overview

From the Settings page, you can configure:

* **IRC**
  * Configure the Network, Nickname and Username information for the bot. You can also enable TLS-- be sure to specify a TLS-capable port for the Network.
  * Enable SASL authentication and enter your NickServ username and password. **Note these are stored plaintext in a blob format.**
  * Owner Nick and Host can be used from within scripts to verify a user's ownership of the bot. See [the JavaScript reference](js-api.md) for more information.
* **Script Management**
  * If you prefer to work with an external editor, you can have squIRCy2 load scripts on the filesystem. Specify a path, enable the option, and restart squIRCy2.
  * Import and Export scripts from the embedded data store.
* **Web Interface**
  * Disable the web interface completely by disabling both HTTP and HTTPS.
  * Configure HTTPS by specifying a certificate file and private key.
  * Configure HTTP(S) Basic Authentication with a Username and Password. **Note these are stored plaintext in a blob format.**
