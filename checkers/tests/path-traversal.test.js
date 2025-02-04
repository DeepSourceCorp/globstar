const fs = require("fs");
const path = require("path");

fs.readFile("../config.json", "utf8", (err, data) => {
  console.log(data);
});

fs.writeFileSync("./users/data.txt", "sensitive data");

path.readdir("../../documents", (err, files) => {
  console.log(files);
});

fs.unlink("../temp/file.txt", (err) => {
  if (err) throw err;
});

fs.access("/users/admin", (err) => {
  console.log(err ? "no access" : "can access");
});
