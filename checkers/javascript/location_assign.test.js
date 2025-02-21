// <expect-error>
location.assign(document.getElementById("input").value);
// <expect-error>
location.assign(new URLSearchParams(window.location.search).get("redirect"));