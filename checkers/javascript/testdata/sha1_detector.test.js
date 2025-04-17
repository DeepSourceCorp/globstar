const jsSHA = require("jssha");

// ok
new jsSHA("SHA-512", "TEXT", { encoding: "UTF8" });

// <expect-error>
new jsSHA("SHA-1", "TEXT", { encoding: "UTF8" });