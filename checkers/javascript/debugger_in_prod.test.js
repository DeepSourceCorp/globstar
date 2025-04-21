if (condition == true){
    x = "y"
}else{
    x = "21"
    //<expect-error>
    debugger;
}

console.log("Five\n");

// <expect-error>
debugger;