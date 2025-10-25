// Test file for no-any checker
// This file should trigger the no-any checker

// <expect-error>
function processData(data: any) {
  return data;
}

// <expect-error>
const handleInput = (input: any): void => {
  console.log(input);
};

class DataProcessor {
  // <expect-error>
  process(item: any) {
    return item;
  }
}

interface Config {
  // <expect-error>
  value: any;
}

//Should NOT be flagged - using unknown
function processUnknown(data: unknown) {
  return data;
}

//Should NOT be flagged - using specific type
function processUser(user: User) {
  return user;
}

//Should NOT be flagged - using generic
function processGeneric<T>(data: T): T {
  return data;
}

interface User {
  name: string;
  age: number;
}
