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
