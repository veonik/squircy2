/**
 * This script is an example of automatic channel joining on connect.
 */
(function setupAutojoin() {
    bind("irc.CONNECT", function(e) {
        Irc.Join("#squishyslab");
    });
})();