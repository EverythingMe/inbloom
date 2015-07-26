package me.everything.inbloom;

import java.io.InvalidObjectException;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Arrays;
import java.util.zip.CRC32;

/**
 * Pure Java Bloom Filter class.
 *
 * Translated from the libbloom C library
 *
 * Original copyright note from libbloom:
 *
 *      Copyright (c) 2012, Jyri J. Virkki
 *      All rights reserved.
 *
 *      This file is under BSD license. See LICENSE file.
*/
public class BloomFilter {

    private static final String TAG = "BloomFilter" ;
    // These fields are part of the public interface of this structure.
    // Client code may read these values if desired. Client code MUST NOT
    // modify any of these.
    int entries;
    double error;
    int bits;
    int bytes;
    int hashes;

    // Fields below are private to the implementation. These may go away or
    // change incompatibly at any moment. Client code MUST NOT access or rely
    // on these.
    double bpe;
    byte[] bf;
    int ready;

    /**
     * Create a blank bloom filter, with a given number of expected entries, and an error rate
     * @param entries the expected number of entries
     * @param error the desired error rate
     */
    public BloomFilter(int entries, double error)
    {
        this(null, entries, error);
    }


    static private final double denom = 0.480453013918201;

    /**
     * Create a bloom filter from an existing data buffer, created by another instance of this library (probably in another language).
     * If the length of the data does not fit the number of entries and error rate, we raise RuntimeException
     * @param data the raw filter data
     * @param entries the expected number of entries
     * @param error the desired error rate
     */
    public BloomFilter(byte []data, int entries, double error) throws RuntimeException
    {
        if (entries < 1 || error == 0) {
            throw new RuntimeException("Invalid params for bloom filter");
        }

        this.entries = entries;
        this.error = error;

        bpe = -(Math.log(error) / denom);
        bits = (int)((double)entries * bpe);
        bytes = (bits / 8) + (bits % 8 != 0 ? 1 : 0);

        if (data != null) {
            if (bytes != data.length) {
                throw new RuntimeException(String.format("Expected %d bytes, got %d", bytes, data.length));
            }
            bf = data;
        } else {
            bf = new byte[bytes];;
        }


        hashes = (int)Math.ceil(0.693147180559945 * bpe);  // ln(2)
    }

    public static short computeChecksum(byte[] data) {
        CRC32 crc = new CRC32();
        crc.update(data);
        long checksum32 = crc.getValue();
        return (short) ((checksum32 & 0xFFFF) ^ (checksum32 >> 16));
    }

    public static BloomFilter load(byte[] bytes) throws InvalidObjectException {
        ByteBuffer bb = ByteBuffer.wrap(bytes);
        bb.order(ByteOrder.BIG_ENDIAN);
        short checksum = bb.getShort();
        short errorRate = bb.getShort();
        int cardinality = bb.getInt();
        final byte[] data = Arrays.copyOfRange(bytes, bb.position(), bytes.length);
        if (computeChecksum(data) != checksum)
            throw new InvalidObjectException("Bad checksum");

        return new BloomFilter(data, cardinality, 1.0 / errorRate);
    }

    public static byte[] dump(BloomFilter bf) {
        // 8 is the size of the header
        byte[] bytes = new byte[bf.bytes + 8];
        ByteBuffer bb = ByteBuffer.wrap(bytes);
        bb.order(ByteOrder.BIG_ENDIAN);

        bb.putShort(computeChecksum(bf.bf));
        bb.putShort((short) (1.0 / bf.error));
        bb.putInt(bf.entries);
        bb.put(bf.bf);

        return bytes;
    }

    private static long unsigned(int i) {
        return i & 0xffffffffl;
    }

    /**
     * check existence or add an entry
     * @param key the key to check/add
     * @param add whether we add or just check the existence
     * @return true if the key is already in the filter
     */
    private boolean checkAdd(String key,boolean add) {

        int hits = 0;
        long a = unsigned(Murmur2.hash32(key, 0x9747b28c));
        long b = unsigned(Murmur2.hash32(key, (int) a));



        for (int i = 0; i < hashes; i++) {
            long x = unsigned ((int)(a + i*b)) % bits;
            long bt = x >> 3;

            byte c = bf[(int)bt];        // expensive memory access
            byte mask = (byte)(1 << (x % 8));

            if ((c & mask) != 0) {
                hits++;
            } else {
                if (add) {
                    bf[(int)bt] = (byte)(c | mask);
                }
            }


        }

        return hits == hashes;
    }


    /**
     * Check whether the filter contains a string
     * @param key the string to check
     * @return true if it already exists in the filter
     */
    public boolean contains(String key)  {
        return checkAdd(key, false);
    }


    /**
     * Add a string to the filter
     * @param key the string to add
     * @return true if the string was already in the filter
     */
    public boolean add(String key) {
        return checkAdd(key, true);
    }
}
