function dangerousEval(userInput) {
  eval(userInput);
}

function dangerousFunction(userInput) {
  const fn = new Function(`return ${userInput}`);
  return fn();
}
