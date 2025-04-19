// <expect-error>
public class MyProprietaryMessageDigest extends MessageDigest {

    @Override
    protected byte[] engineDigest() {
        return "";
    }
}

// <no-error>
public class NotMessageDigest {
    public NotMessageDigest() {
        System.out.println("");
    }
}
