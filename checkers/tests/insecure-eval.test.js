let userInput = "Hello World";
eval("alert('" + userInput + "')");

let calculation = "2 + 2";
Function("return " + calculation)();

let result = eval("10 * 5");
console.log(result);

const dynamicFunction = new Function("a", "b", "return a + b");
dynamicFunction(5, 3);
