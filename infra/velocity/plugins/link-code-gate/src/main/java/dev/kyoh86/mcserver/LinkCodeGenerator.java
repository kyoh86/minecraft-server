package dev.kyoh86.mcserver;

import java.security.SecureRandom;

final class LinkCodeGenerator {
  private static final String ALPHABET = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ";
  private static final int CODE_LENGTH = 8;

  private final SecureRandom random;

  LinkCodeGenerator(SecureRandom random) {
    this.random = random;
  }

  String generate() {
    StringBuilder sb = new StringBuilder(CODE_LENGTH);
    for (int i = 0; i < CODE_LENGTH; i++) {
      int idx = random.nextInt(ALPHABET.length());
      sb.append(ALPHABET.charAt(idx));
    }
    return sb.toString();
  }
}
