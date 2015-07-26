package me.everything.inbloom;

/**
 * Utility class for dealing with binary data
 */
public class BinAscii {
    final protected static char[] hexArray = "0123456789ABCDEF".toCharArray();

    /**
     * Transform a byte array into a it's hexadecimal representation
     */
    public static String hexlify(byte[] bytes) {
        char[] hexChars = new char[bytes.length * 2];
        for ( int j = 0; j < bytes.length; j++ ) {
            int v = bytes[j] & 0xFF;
            hexChars[j * 2] = hexArray[v >>> 4];
            hexChars[j * 2 + 1] = hexArray[v & 0x0F];
        }
        String ret = new String(hexChars);
        return ret;
    }

    /**
     * Transform a string of hexadecimal chars into a byte array
     */
    public static byte[] unhexlify(String argbuf) {
        int arglen = argbuf.length();
        if (arglen % 2 != 0)
            throw new RuntimeException("Odd-length string");

        byte[] retbuf = new byte[arglen/2];

        for (int i = 0; i < arglen; i += 2) {
            int top = Character.digit(argbuf.charAt(i), 16);
            int bot = Character.digit(argbuf.charAt(i+1), 16);
            if (top == -1 || bot == -1)
                throw new RuntimeException("Non-hexadecimal digit found");
            retbuf[i / 2] = (byte) ((top << 4) + bot);
        }
        return retbuf;
    }
}
