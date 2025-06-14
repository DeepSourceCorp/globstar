function getUserInput(key) {

    return document.getElementById(key).value;

}

userInput = getUserInput('username')

// A sink method, which performs some raw databse operation on the userInput
// <expect-error>
perform_db_operation(userInput)