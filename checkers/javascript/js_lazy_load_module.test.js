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
    mod = require('some-module');
    return mod.doSomn();
}

const arrowFunction = () => {
    //<expect-error>
    const someModule = require('some-module');
    return someModule.doSomn();
}


const expressionFunction = function() {
    //<expect-error> 
    const anotherModule = require('some-module');
    return anotherModule.doSomething();
};


function outerFunction() {
    function innerFunction() {
        //<expect-error>
        const yetAnotherModule = require('some-module');
        return yetAnotherModule.doSomething();
    }
  return innerFunction();
}
