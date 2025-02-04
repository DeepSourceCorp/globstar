const MongoDB = require("mongodb");
const mongoose = require("mongoose");

// Vulnerable queries
async function unsafeQueries(userInput) {
  // Direct string interpolation in query
  // <expect-error>
  const user1 = await db.collection("users").findOne({
    username: `${userInput}`,
  });

  // Template literal in query
  // <expect-error>
  const user2 = await db.collection("users").find({
    email: `${userInput.email}`,
  });

  // Multiple conditions with injection
  // <expect-error>
  await db.collection("users").update({
    username: `${userInput.username}`,
    password: `${userInput.password}`,
  });

  // Delete operation
  // <expect-error>
  await db.collection("posts").remove({
    author: `${userInput.author}`,
  });
}

// Safe queries
async function safeQueries(userInput) {
  // Using proper sanitization and type checking
  const sanitizedInput = validateAndSanitize(userInput);

  // Using MongoDB operators
  const user = await db.collection("users").findOne({
    username: { $eq: sanitizedInput },
  });

  // Using Mongoose schema validation
  const UserModel = mongoose.model("User");
  const result = await UserModel.findOne({
    email: sanitizedInput.email,
  });

  // Using parameterized values
  const query = { username: 1 };
  const result2 = await db.collection("users").find(query);
}

// Example of sanitization
function validateAndSanitize(input) {
  if (typeof input !== "string") {
    throw new Error("Invalid input type");
  }
  return input.replace(/[^a-zA-Z0-9]/g, "");
}

// Examples of MongoDB injection attacks
const maliciousQueries = {
  // Object injection
  username: { $ne: null },

  // Operator injection
  password: { $exists: true },

  // Regex injection
  email: { $regex: ".*" },

  // Function injection
  age: { $where: "function() { return true; }" },
};

// <expect-error>
db.collection("users").find({
  username: `${maliciousQueries.username}`,
});
