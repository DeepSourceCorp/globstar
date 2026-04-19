// Test file for no-non-null-assertion checker

interface User {
  name: string;
  email?: string;
}

function getUserEmail(user: User | null) {
  // <expect-error>
  return user!.email;
}

// <expect-error>
const element = document.getElementById("myId")!;

// <expect-error>
const value = possiblyUndefined!.property;

function process(data: string | undefined) {
  // <expect-error>
  const length = data!.length;
  return length;
}

//Should NOT be flagged - using optional chaining
function getUserEmailSafe(user: User | null) {
  return user?.email;
}

//Should NOT be flagged - using null check
const elementSafe = document.getElementById("myId");
if (elementSafe) {
  elementSafe.classList.add("active");
}

//Should NOT be flagged - using nullish coalescing
function processSafe(data: string | undefined) {
  const length = data?.length ?? 0;
  return length;
}
