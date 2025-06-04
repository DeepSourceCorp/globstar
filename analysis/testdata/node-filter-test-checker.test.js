console.log("Hello, world!");

function foo(){
    // <expect-error>
    console.log("This should be detected");

    /*
    console.log("This Should not be detected");
    */
}