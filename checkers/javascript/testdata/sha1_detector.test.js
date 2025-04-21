import shapkg from "jssha"

{
    const jsSHA = require("jssha");

    // ok
    new jsSHA("SHA-512", "TEXT", { encoding: "UTF8" });

    // <expect-error>
    new jsSHA("SHA-1", "TEXT", { encoding: "UTF8" });

}

new shapkg("SHA-512", "TEXT", { encoding: "UTF8" })

// <expect-error>
new shapkg("SHA-1", "TEXT", { encoding: "UTF8" })