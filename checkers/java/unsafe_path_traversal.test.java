public class UnsafePathTraversalTest {

  // These should be flagged
  public void testVulnerable() {
    String userInput = request.getParameter("file");

    // <expect-error>
    File file1 = new File(userInput);

    // <expect-error>
    File file2 = new File("/var/data/" + userInput + "test");

    // <expect-error>
    File file3 = new File(String.format("/var/www/uploads/%s", userInput));
  }

  // These should not be flagged
  public void testSafe() {
    String userInput = request.getParameter("file");

    // path is commented
    // File file3 = new File(userInput);

    // Safe usage: constant file path used
    File file4 = new File("/var/data/file.txt");
  }
}
