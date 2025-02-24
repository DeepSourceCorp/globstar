<?php

function test_dangerous_eval() {
    $user_input = $_GET['input'];

    // These should be flagged
    // <expect-error>
    eval($user_input);

    // <expect-error>
    eval("echo " . $user_input . "hi");

    // String interpolation
    // <expect-error>
    eval("echo $user_input");

    // Superglobal (outside our control) sources
    // <expect-error>
    eval($_GET['username']);

    // These are safe and should not be flagged
    // constants
    eval('echo "Hello, World!"');

}

function test_edge_cases() {
    // Should not flag eval in variable names
    $evaluation_result = 100;

    // Should not flag commented-out eval
    // eval($user_input);
}
