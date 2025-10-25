// Test file for no-explicit-any checker

// <expect-error>
function parseJson(json: string): any {
  return JSON.parse(json);
}

// <expect-error>
const processDataSafer = (data: any): void => {
  console.log(data);
};

interface Config {
  // <expect-error>
  value: any;
  // <expect-error>
  settings: any;
}

class DataHandler {
  // <expect-error>
  handle(input: any) {
    return input;
  }
}

//Should NOT be flagged - using unknown
function parseJsonSafe(json: string): unknown {
  return JSON.parse(json);
}

//Should NOT be flagged - using generics
function parseJsonGeneric<T>(json: string): T {
  return JSON.parse(json);
}

//Should NOT be flagged - using specific type
interface User {
  name: string;
  email: string;
}

function saveUser(user: User): void {
  console.log(user.name);
}
