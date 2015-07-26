package me.everything.inbloom;

import junit.framework.TestCase;

import java.io.InvalidObjectException;

/**
 * Created by dvirsky on 23/07/15.
 */
public class BloomFilterTest extends TestCase {

    public void testFilter() throws InvalidObjectException {
        BloomFilter bf = new BloomFilter(20, 0.01);


        bf.add("foo");
        bf.add("bar");
        bf.add("foosdfsdfs");
        bf.add("fossdfsdfo");
        bf.add("foasdfasdfasdfasdfo");
        bf.add("foasdfasdfasdasdfasdfasdfasdfasdfo");


        assertTrue(bf.contains("foo"));
        assertTrue(bf.contains("bar"));

        assertFalse(bf.contains("baz"));
        assertFalse(bf.contains("faskdjfhsdkfjhsjdkfhskdjfh"));


        BloomFilter bf2 = new BloomFilter(bf.bf, bf.entries, bf.error);
        assertTrue(bf2.contains("foo"));
        assertTrue(bf2.contains("bar"));

        assertFalse(bf2.contains("baz"));
        assertFalse(bf2.contains("faskdjfhsdkfjhsjdkfhskdjfh"));

        String serialized = BinAscii.hexlify(BloomFilter.dump(bf));
        System.out.printf("Serialized: %s\n", serialized);

        String hexPayload = "620d006400000014000000000020001000080000000000002000100008000400";
        BloomFilter deserialized = BloomFilter.load(BinAscii.unhexlify(hexPayload));
        String dump = BinAscii.hexlify(BloomFilter.dump(deserialized));
        System.out.printf("Re-Serialized: %s\n", dump);
        assertEquals(dump.toLowerCase(), hexPayload);

        //BloomFilter deserialized = BloomFilter.load(BloomFilter.dump(bf));


        assertEquals(deserialized.entries, 20);
        assertEquals(deserialized.error, 0.01);
        assertTrue(deserialized.contains("abc"));
    }
}