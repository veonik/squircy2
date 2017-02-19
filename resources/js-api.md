Javascript API
==============

squIRCy2 embeds [otto](https://github.com/robertkrimen/otto) and supports ECMAScript 5, minus a regular expression incompatibility.
Additionally, the following functions are available to interact with the various squIRCy2 modules:

| Method | Description |
| ------ | ----------- |
| Irc.Join(channel) | Joins the given channel |
| Irc.Part(channel) | Parts the given channel |
| Irc.Privmsg(target, message) | Messages target with message. Target can be a user or a channel |
| Irc.CurrentNick() | Get the bot's current nickname |
| Irc.Nick(newNick) | Change the bot's nickname |
| Http.Get(url, ...headers) | Fetch the given url using a GET HTTP request |
| Http.Post(url, body, ...headers) | Fetch the given url using a POST HTTP request |
| Http.Send(options) | Send an HTTP request with the configured options. |
| Math.Rand() | Generate a random value from 0-1 |
| Math.Round(val) | Round val to 0 decimal places. |
| Config.OwnerNick() | Get the configured Owner Nickname |
| Config.OwnerHost() | Get the configured Owner Host |
| bind(eventName, fnName) | Add a handler of the given event type and function name |
| unbind(eventName, fnName) | Removes a handler of the given type and function name |
| setTimeout(fnName, delay) | Executes fnName after delay milliseconds |
| setInterval(fnName, delay) | Executes fnName every delay milliseconds |
| use(coll) | Opens and returns a repository for the given collection |

## Repository methods

These are methods available on a repository returned by `use`.

| Method | Description |
| ------ | ----------- |
| repo.Fetch(id) | Attempts to load and return an entity with the given id |
| repo.FetchAll() | Returns a collection of all the entities in the repository |
| repo.Save(entity) | Saves the given entity |

## Stubs

If you're looking for IDE auto-completion for your squIRCy2 scripts, you can copy the [squircy stubs](stubs/squircy.js) file into your configured squIRCy2 scripts directory.

## Event Handling

See the [dedicated section](event-handling.md) for details.