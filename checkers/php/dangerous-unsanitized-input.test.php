<?php
// XSS Vulnerability Test File
function test_echo_direct_unsanitized()
{
    // Vulnerable: User input is output directly.
    // <except-error>
    echo $_GET["input"];
}

// function test_echo_concatenation()
// {
//     // Vulnerable: Concatenates unsanitized input into a greeting.
//     // <except-error>
//     echo "Hello, " . $_POST["username"] . "!";
// }

function test_print_unsanitized()
{
    // Vulnerable: Directly prints unsanitized user data.
    // <except-error>
    print $_REQUEST["data"];
}

function test_html_attribute_injection()
{
    // Vulnerable: Unsanitized URL injected into an anchor tag.
    // <except-error>
    echo '<a href="' . $_GET["url"] . '">Click Here</a>';
}

function test_input_field_unsanitized()
{
    // Vulnerable: The value attribute is populated without encoding.
    // <except-error>
    echo '<input type="text" value="' . $_GET["value"] . '">';
}

function test_script_block_unsanitized()
{
    // Vulnerable: User input is embedded directly in a script.
    // <except-error>
    echo "<script>";
    echo 'var userInput = "' . $_GET["jsdata"] . '";';
    echo "alert(userInput);";
    echo "</script>";
}

function test_double_quote_array_access()
{
    // Vulnerable: Functionally similar to single quotes; unsanitized input is echoed.
    // <except-error>
    echo $_GET["param"];
}

function test_dom_manipulation()
{
    // Vulnerable: User content is injected into both HTML and JavaScript contexts.
    // <except-error>
    echo '<div id="output">' . $_GET["content"] . "</div>";
    echo '<script>document.getElementById("output").innerHTML = "' .
        $_GET["content"] .
        '";</script>';
}

// Safer usage of echo and print
function test_safe_echo()
{
    // no error should be reported
    echo htmlspecialchars($_GET["input"], ENT_QUOTES, "UTF-8");
}

function test_safe_filter_var()
{
    // no error should be reported
    $safe_url = filter_var($_GET["url"], FILTER_SANITIZE_URL);
    echo '<a href="' .
        htmlspecialchars($safe_url, ENT_QUOTES, "UTF-8") .
        '">Click Here</a>';
}

function test_safe_json_encode()
{
    // no error should be reported
    echo "<script>";
    echo "var userInput = " . json_encode($_GET["jsdata"]) . ";";
    echo "alert(userInput);";
    echo "</script>";
}
?>
