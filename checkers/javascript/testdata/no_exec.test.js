let x = window.prompt()

function foo(x) {
   exec(x)
}

// <expect-error>
foo(x)