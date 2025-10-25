// Test file for ts-console-in-prod checker

function processDataWithConsole(data: any[]) {
  // <expect-error>
  console.log("Processing data:", data);
  return data;
}

class UserService {
  login(credentials: { username: string; password: string }) {
    // <expect-error>
    console.log("Login attempt:", credentials);
    // ... login logic
  }
}

const debugInfo = (info: string) => {
  // <expect-error>
  console.debug(info);
};

try {
  // some code
  throw new Error("test");
} catch (error) {
  // <expect-error>
  console.error("Error occurred:", error);
}

// ✅ Should NOT be flagged - using proper logger
import logger from "./logger";

function processDataSafe(data: any[]) {
  logger.info("Processing data:", data);
  return data;
}
