const crypto = require("crypto");
const CryptoJS = require("crypto-js");
const md5 = require("md5");
const sha1 = require("sha1");

const password = "mysecretpassword123";

const hash1 = crypto.MD5(password);
const hash2 = CryptoJS.SHA1(password);
const hash3 = md5(password);
const hash4 = sha1(password);

console.log("MD5 hash:", hash1);
console.log("SHA1 hash:", hash2);
console.log("Another MD5:", hash3);
console.log("Another SHA1:", hash4);
