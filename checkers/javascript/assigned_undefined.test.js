// 'undefined' is "assignable" syntactically but it's read-only (since
// ECMAScript 5), so its value will remain 'undefined'.
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/undefined

// Ok.
alert(undefined); //alerts "undefined"

// <expect-error> 
var undefined = "new value";
alert(undefined) // alerts "new value"

// <expect-error>
undefined = "new value";
alert(undefined) // alerts "new value"
