// Match
// <expect-error>
location.assign(document.getElementById("input").value);
// <expect-error>
location.assign(new URLSearchParams(window.location.search).get("redirect"));
// <expect-error>
window.location.assign(document.getElementById("input").value);
// <expect-error>
window.location.assign(new URLSearchParams(window.location.search).get("redirect"));