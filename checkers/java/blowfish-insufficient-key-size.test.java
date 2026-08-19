public class Cls {

    public void unsafeKeySize() {
        KeyGenerator keyGen = KeyGenerator.getInstance("Blowfish");
        // <expect-error>
        keyGen.init(64);
    }

    public void safeKeySize() {
        // ok
        KeyGenerator keyGen = KeyGenerator.getInstance("Blowfish");
        keyGen.init(128);
    }

    public void superSafeKeySize() {
        // ok
        KeyGenerator keyGen = KeyGenerator.getInstance("Blowfish");
        keyGen.init(448);
    }

    public void invalidKeySize() {
        // ok
        KeyGenerator keyGen = KeyGenerator.getInstance("Blowfish");
        keyGen.init(-1);
    }
}