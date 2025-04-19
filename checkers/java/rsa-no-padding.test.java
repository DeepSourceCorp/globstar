class RSAPadding {
  public void rsaNoPadding() {
    // <expect-error>
    Cipher.getInstance("RSA/NONE/NoPadding");
  }

  public void rsaNoPadding2() {
    // <expect-error>
    useCipher(Cipher.getInstance("RSA/None/NoPadding"));
  }

  public void rsaPadding() {
    // <no-error>
    Cipher.getInstance("RSA/ECB/OAEPWithMD5AndMGF1Padding");
  }
}
