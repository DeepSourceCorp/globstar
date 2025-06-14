function getUserInput(key) {

    return document.getElementById(key).value;

}

userInput = getUserInput('username')

// <expect-error>
perform_db_operation(userInput)