const input = req.query.input;

// ok
eval("alert");

// <expect-error>
eval(input);

// <expect-error>
var x = new Function("a", "b", `return ${input}(a,b)`);

// <expect-error>
var y = Function("a", "b", input);

setTimeout(() => {
  // ok
  console.log("Delayed for 1 second." + input);
}, 1000);

setTimeout(function () {
  // ok
  console.log("Delayed for 1 second." + input);
}, 1000);

// <expect-error> setTimeout makes the browser do an eval() under the hood with the argument passed
setTimeout("console.log(" + input + ")", 1000);
