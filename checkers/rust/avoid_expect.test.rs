fn test_antipattern() {
    let some_option: Option<i32> = None;

    // <expect-error>
    let _ = some_option.expect("failed to get value");

    let some_result: Result<i32, &str> = Err("Error occurred");

    // <expect-error>
    let _ = some_result.expect("failed to process result");
}

fn test_safe() {
    let some_option: Option<i32> = None;

    // Safe error handling with unwrap_or
    let _ = some_option.unwrap_or(42);

    // Safe error handling with match
    let _ = match some_option {
        Some(val) => val,
        None => 0,
    };

    // Should not flag string literals containing "expect"
    let message = "We expect proper error handling in production code.";
}
