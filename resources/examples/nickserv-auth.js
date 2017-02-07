/**
 * This script enables squIRCy2 to automatically identify with NickServ.
 */
(function setupNickserv() {
    var handled = false;
    bind("irc.NOTICE", function(e) {
        if (!handled && e.Nick == "NickServ" && e.Message.indexOf("identify") >= 0) {
            Irc.Privmsg("NickServ", "IDENTIFY myuser areallysecurepassword");
            console.log("Identified with Nickserv");
            handled = true;
        }
    });
    bind("irc.DISCONNECT", function(e) {
        handled = false;
    });
})();
