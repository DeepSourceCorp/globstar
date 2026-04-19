// Test file for no-eval checker

function dangerousCode(input: string) {
  // <expect-error>
  eval(input);
}

// <expect-error>
const result = eval("2 + 2");

// <expect-error>
const calculate = (expr: string) => eval(expr);

function processConfig() {
  const config = '{ "debug": true }';
  // <expect-error>
  return eval(`(${config})`);
}

//Should NOT be flagged - using JSON.parse
function safeJsonParse(json: string) {
  return JSON.parse(json);
}

//Should NOT be flagged - using function map
const operations = {
  add: (a: number, b: number) => a + b,
  subtract: (a: number, b: number) => a - b,
};

function calculate_something(op: string, a: number, b: number) {
  const operation = operations[op as keyof typeof operations];
  return operation ? operation(a, b) : 0;
}

//Should NOT be flagged - configuration object
const config = {
  development: { apiUrl: "http://localhost:3000" },
  production: { apiUrl: "https://api.example.com" },
};

const environment: "development" | "production" = "development";
const settings = config[environment];
