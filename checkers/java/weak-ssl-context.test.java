import java.lang.Runtime;

class Cls {

    public Cls() {
        System.out.println("Hello");
    }

    public void test1() {
        // <expect-error>
        SSLContext ctx = SSLContext.getInstance("SSL");
    }

    public void test2() {
        // <expect-error>
        SSLContext ctx = SSLContext.getInstance("TLS");
    }

    public void test3() {
        // <expect-error>
        SSLContext ctx = SSLContext.getInstance("TLSv1");
    }

    public void test4() {
        // <expect-error>
        SSLContext ctx = SSLContext.getInstance("SSLv3");
    }

    public void test5() {
        // <expect-error>
        SSLContext ctx = SSLContext.getInstance("TLSv1.1");
    }

    public void test6() {
        // <no-error>
        SSLContext ctx = SSLContext.getInstance("TLSv1.2");
    }

    public void test7() {
        // <no-error>
        SSLContext ctx = SSLContext.getInstance("TLSv1.3");
    }

    public String getSslContext() {
        return "Anything";
    }

    public void test8() {
        // <no-error>
        SSLContext ctx = SSLContext.getInstance(getSslContext());
    }
}
