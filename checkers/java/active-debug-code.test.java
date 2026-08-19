public class ActiveDebugCode{

    public void bad(){
        StackTraceElement[] elements;

        Exception e = new Exception();
        elements = e.getStackTrace();

        // <expect-error>
        System.err.print(elements);
    }

    public void bad2(){
        StackTraceElement[] elements;
        elements = Thread.currentThread().getStackTrace();

        // <expect-error>
        System.err.print(elements);
    }

    public void bad3(){
        StackTraceElement[] elements;
        elements = new Throwable().getStackTrace();

        // <expect-error>
        System.err.print(elements);
    }

    public void bad4(){
        // <expect-error>
        System.out.println(org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e));
        // <expect-error>
        System.out.println(org.apache.commons.lang3.exception.ExceptionUtils.getFullStackTrace(e));
    }

    public void alsobad(){
        for (StackTraceElement ste : Thread.currentThread().getStackTrace()) {
            // <expect-error>
            System.out.println(ste);
        }
    }

}

