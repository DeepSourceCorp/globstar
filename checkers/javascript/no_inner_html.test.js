

// match
document.body.innerHTML = `${name}`
document.body.innerHTML += document.getElementById("name").value;
document.querySelector("#div1").innerHTML += `${name}`;
document.querySelector("#div2").innerHTML += document.getElementById("name").value;
article.innerHTML += `${name}`;
article.innerHTML += document.getElementById("name").value;
// Dont match 
document.body.innerHTML += "<p>Hello</p>";
document.querySelector("#div3").innerHTML = DOMPurify.sanitize(userInput);
document.querySelector("#div4").innerHTML = DOMPurify.sanitize(document.getElementById("name").value);
article.innerHTML = "<p>Hello</p>";
article.innerHTML = DOMPurify.sanitize(userInput);