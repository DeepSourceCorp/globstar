unsafe fn dangerous() {}
// <expect-error>
unsafe {
    dangerous();
}