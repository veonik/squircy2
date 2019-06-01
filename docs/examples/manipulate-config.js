/**
 * This script shows how to access and manipulate the full runtime configuration.
 */
(function(window) {
  var repo = use("Settings");
  
  var conf = repo.FetchAll()[0];
  if (!conf) {
    console.log("expected conf to be non-empty!");
    return;
  }
  
  // Read the sensitive information, no big deal
  console.log("your AuthPassword is: " + conf.AuthPassword);
  
  // Change one of the settings, then save the changes.
  conf.Username = "newuser";
  repo.Save(conf);
  // Check the web UI to confirm the change
  
  // Print the changes
  console.log("the bots new IRC Username (on next connect) is: " + conf.Username);  
})(this);
