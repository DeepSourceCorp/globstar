fn test_antipattern() {
    // These should be flagged
    let some_option: Option<i32> = None;

    // <expect-error>
    let _ = some_option.unwrap();

    let some_result: Result<i32, &str> = Err("Error occurred");

    // <expect-error>
    let _ = some_result.unwrap();

}

fn test_safe() {}
    // These are safe and should not be flagged

    // Handling Option safely
    let _ = some_option.unwrap_or(42); // Providing a default value

    // some_option.unwrap();

    // Should not flag string literals containing "unwrap"
    let message = "Do not use unwrap in production!";
}
