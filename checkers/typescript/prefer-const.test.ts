// Test file for prefer-const checker

// <expect-error>
let userName = "John";
console.log(userName);

// <expect-error>
let config = {
  api: "https://api.example.com",
  timeout: 5000,
};

// <expect-error>
let numbers = [1, 2, 3, 4, 5];

// <expect-error>
let settings = { debug: false };
settings.debug = true; // mutation is ok, reassignment is not

//Should NOT be flagged - using const
const userName2 = "Jane";
console.log(userName2);

// NOTE: This checker has limitations - it flags all 'let' declarations
// In production, you'd want scope analysis to check reassignment
// <expect-error>
let counter = 0;
counter = 1;
counter += 1;

// <expect-error>
for (let i = 0; i < 10; i++) {
  console.log(i);
}
