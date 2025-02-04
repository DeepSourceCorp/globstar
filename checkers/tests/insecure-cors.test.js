// Express example
const express = require("express");
const app = express();

app.use((req, res, next) => {
  // <expect-error>
  res.header("Access-Control-Allow-Origin", "*");
  res.header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE");
  next();
});

// Direct response headers
function handleRequest(response) {
  // <expect-error>
  response.setHeader("Access-Control-Allow-Origin", "*");
  response.setHeader("Content-Type", "application/json");
}

// Config object
const corsConfig = {
  // <expect-error>
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET,POST",
};

// Raw headers object
const headers = {};
// <expect-error>
headers["Access-Control-Allow-Origin"] = "*";

// Server configuration
const serverConfig = {
  cors: {
    // <expect-error>
    "Access-Control-Allow-Origin": "*",
  },
};
