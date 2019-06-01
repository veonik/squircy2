/**
 * This script enables squIRCy2 to maintain a list of users in each of the
 * channels it joins.
 */
(function(Irc) {

    /**
     * A map of channel names to a list of their participant nicknames.
     */
    Irc.Channels = {};
    /**
     * Processes a NAMES command for the given channel.
     * @param {string} channel
     */
    Irc.Names = function(channel) {
        var users = [];
        var listen = function(e) {
            var chan = e.Args[2];
            if (chan !== channel) {
                return;
            }
            var nicks = e.Args[3].replace(/[^a-z0-9_ ]/gi, '').split(" ");
            users = users.concat(nicks);
        }
        var end = function() {
            unbind("irc.353", listen);
            unbind("irc.366", end);
            Irc.Channels[channel] = users;
        }
        bind("irc.353", listen);
        bind("irc.366", end);
        Irc.Raw("NAMES :"+channel);
    }

    bind("irc.JOIN", function(e) {
        Irc.Names(e.Target);
    });

    bind("irc.PART", function(e) {
        Irc.Names(e.Target);
    });

})(Irc);