class ECBCipher {

  public void ecbCipher() {
    // <expect-error>
    Cipher c = Cipher.getInstance("AES/ECB/NoPadding");
    c.init(Cipher.ENCRYPT_MODE, k, iv);
    byte[] cipherText = c.doFinal(plainText);
  }
  public void noEcbCipher() {
    // <no-error>
    Cipher c = Cipher.getInstance("AES/GCM/NoPadding");
    c.init(Cipher.ENCRYPT_MODE, k, iv);
    byte[] cipherText = c.doFinal(plainText);
  }
}
