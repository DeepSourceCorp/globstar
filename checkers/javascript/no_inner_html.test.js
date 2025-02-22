

// Match
// <expect-error>
document.body.innerHTML = `${name}`
// <expect-error>
document.body.innerHTML += document.getElementById("name").value;
// <expect-error>
document.querySelector("#div1").innerHTML += `${name}`;
// <expect-error>
document.querySelector("#div2").innerHTML += document.getElementById("name").value;
// <expect-error>
article.innerHTML += `${name}`;
// <expect-error>
article.innerHTML += document.getElementById("name").value;

// Dont match 
// <no-error>
document.body.innerHTML += "<p>Hello</p>";
// <no-error>
document.querySelector("#div3").innerHTML = DOMPurify.sanitize(userInput);
// <no-error>
document.querySelector("#div4").innerHTML = DOMPurify.sanitize(document.getElementById("name").value);
// <no-error>
article.innerHTML = "<p>Hello</p>";
// <no-error>
article.innerHTML = DOMPurify.sanitize(userInput);