// <expect-error>
confirm("Basic");

// <expect-error>
window.confirm("basic as well")

// <expect-error>
if (confirm("pushed!") == true){
    r = 'x';
}else {
    r = 'y';
}