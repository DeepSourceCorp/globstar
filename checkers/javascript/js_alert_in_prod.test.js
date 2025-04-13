//<expect-error>
alert("legacy code supremacy");

function triggerBasicAlert() {
    //<expect-error>
    alert("This is an alert");
}

document.getElementById("button").onclick = function() {
    //<expect-error>
    alert("You clicked the button");
};

//<expect-error>
window.alert("if it broke, don't fix it")