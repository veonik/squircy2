/**
 * This file contains stubs for the squIRCy2 JavaScript runtime.
 *
 * Copy this file into your configured squIRCy2 scripts directory to enable
 * IDE auto-completion for squIRCy2.
 */

/**
 * Provides interaction with the IRC module.
 */
var Irc = {
    /**
     * Joins the given channel.
     * @param {string} channel
     */
    Join: function(channel) {},

    /**
     * Parts the given channel.
     * @param {string} channel
     */
    Part: function(channel) {},

    /**
     * Messages target with message. Target can be a user or a channel.
     * @param {string} target
     * @param {string} message
     */
    Privmsg: function(target, message) {},

    /**
     * Get the bot's current nickname.
     * @returns {string}
     */
    CurrentNick: function() {},

    /**
     * Change the bot's nickname.
     * @param {string} newNick
     */
    Nick: function(newNick) {},

    /**
     * Send a raw command.
     * @param {string} raw
     */
    Raw: function(raw) {}
};

var Http = {
    /**
     * Fetch the given url using a GET HTTP request.
     * @param {string}    url
     * @param {...string} headers
     */
    Get: function(url, headers) {},

    /**
     * Fetch the given url using a POST HTTP request.
     * @param {string}   url
     * @param {string}   body
     * @param {...string} headers
     */
    Post: function(url, body, headers) {},

    /**
     * Send an HTTP request with the configured options.
     * @param {object} options
     */
    Send: function(options) {}
};

var Math = {
    /**
     * Generates a random number between 0 and 1.
     * @returns {number}
     */
    Rand: Math.random,

    /**
     * Rounds the given value to 0 decimal places.
     * @param {number}
     */
    Round: Math.round,

    /**
     * Rounds the given value up to 0 decimal places.
     * @param {number}
     */
    Ceil: Math.ceil,

    /**
     * Rounds the given value down to 0 decimal places.
     * @param {number}
     */
    Floor: Math.floor
};

var Config = {
    /**
     * Returns the configured Owner Nickname.
     * @returns {string}
     */
    OwnerNick: function() {},

    /**
     * Returns the configured Owner Hostname.
     * @returns {string}
     */
    OwnerHost: function() {}
};

/**
 * Add a handler of the given event type and function name.
 * @param {string}          eventName
 * @param {Function|string} fnName
 * @return {string} A reference that may be used later in calls to unbind.
 */
function bind(eventName, fnName) {}

/**
 * Removes a handler of the given type and function name.
 * @param {string}          eventName
 * @param {Function|string} fnName
 */
function unbind(eventName, fnName) {}

/**
 * Opens a collection object for the given coll name.
 * @param {string} coll The name of the collection to open.
 * @returns {{FetchAll: FetchAll, Fetch: Fetch, Save: Save, Delete: Delete}}
 */
function use(coll) {
    return {
        /**
         * Fetches all records in the collection.
         * @returns {Array}
         */
        FetchAll: function() {},

        /**
         * Fetches a single record in the collection.
         * @param {number} id
         * @returns {Object}
         */
        Fetch: function(id) {},

        /**
         * Save a record in the collection.
         * @param {Object} entity
         */
        Save: function(entity) {},

        /**
         * Delete a given record in the collection.
         * @param {number} id
         */
        Delete: function(id) {}
    }
}

