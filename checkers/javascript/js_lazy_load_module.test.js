// ok
const fs = require('fs')

function smth() {
  // <expect-error>
  const mod = require('module-name');
  return mod();
}

function smth2(){
    console.log("Hello");
    // <expect-error>
    let mod = require('some-module');
    return mod.doSomn();
}


