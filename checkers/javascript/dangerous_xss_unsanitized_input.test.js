function test_dangerous_xss_unsanitized_input() {
  // These should be flagged for XSS vulnerabilities
  let user_input = getUserInput();

  // <expect-error>
  document.body.innerHTML = user_input;

  // <expect-error>
  document.write(user_input);

  // <expect-error>
  document.body.innerHTML = "<div>" + user_input + "</div>";

  // These are safe and should not be flagged
  let safe_input = "Safe content";
  document.body.innerHTML = safe_input; // No user input involved

  // This should be flagged even inside a function definition
  function dangerousFunction() {
    // <expect-error>
    document.body.innerHTML = getUserInput();
  }

  try {
    // This should not be flagged because it's inside a catch block
    document.body.innerHTML = user_input;
  } catch (err) {
    console.error(err);
  }
}

function getUserInput() {
  return "<script>alert('XSS!')</script>";
}

const htmlContent = `<div>${userInput}</div>`;
// <expect-error>
document.getElementById("output").innerHTML = htmlContent;

function test_dangerous_dom_operations() {
  const userInput = getUserInput();
  const element = document.getElementById("content");

  // These should be flagged

  // <expect-error>
  element.innerHTML = userInput;

  // <expect-error>
  element.innerHTML = "<div>" + userInput + "</div>";

  // <expect-error>
  element.insertAdjacentHTML("beforeend", `${userInput}`);

  // These are safe and should not be flagged

  // Safe because `sanitizeHTML()` sanitizes `userInput` before insertion
  document.getElementById("output").innerHTML = sanitizeHTML(userInput);

  // Safe because there's no user input involved
  document.getElementById("output").innerHTML = "<p>Safe Content</p>";

  // Safe - using textContent
  element.textContent = userInput;

  // Safe - using createElement
  const div = document.createElement("div");
  div.textContent = userInput;
  element.appendChild(div);

  // Safe - using static HTML
  element.innerHTML = "<div>Static content</div>";
}

function test_edge_cases() {
  const element = document.querySelector(".content");

  // Should not flag property access
  const currentHTML = element.innerHTML;

  // Should not flag non-HTML string concatenation
  const message = "Hello, " + username;

  // Should not flag commented code
  // element.innerHTML = userInput;
}

// Helper function to simulate user input
function getUserInput() {
  return "user provided content";
}
