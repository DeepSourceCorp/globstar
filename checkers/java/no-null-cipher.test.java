import java.lang.Runtime;

class Cls {

    public Cls() {
        System.out.println("Hello");
    }

    public byte[] test1(String plainText) {
        // <expect-error>
        javax.crypto.NullCipher nullCipher = new javax.crypto.NullCipher();
        // <expect-error>
        Cipher doNothingCihper = new NullCipher();
        //The ciphertext produced will be identical to the plaintext.
        byte[] cipherText = doNothingCihper.doFinal(plainText);
        return cipherText;
    }

    public void test2(String plainText) {
        // <no-error>
        Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
        byte[] cipherText = cipher.doFinal(plainText);
        return cipherText;
    }

    public void test3(String plainText) {
        // <expect-error>
        useCipher(new NullCipher());
    }

    private static void useCipher(Cipher cipher) throws Exception {
       // sast should complain about the hard-coded key
       SecretKey key = new SecretKeySpec("secret".getBytes("UTF-8"), "AES");
       cipher.init(Cipher.ENCRYPT_MODE, key);
       byte[] plainText  = "aeiou".getBytes("UTF-8");
       byte[] cipherText = cipher.doFinal(plainText);
       System.out.println(new String(cipherText));
    }
}
