package me.everything.inbloom;

import junit.framework.TestCase;

import java.io.InvalidObjectException;

/**
 * Created by dvirsky on 23/07/15.
 */
public class BloomFilterTest extends TestCase {

    public void testCreateFilterWithBadParameters()
    {
        try 
        {
            BloomFilter_mutated bf = new BloomFilter_mutated( -1, 0.01 );
            fail( "should have thrown an exception" );
        }
        catch( RuntimeException e ) 
        {
            assertEquals( "Invalid params for bloom filter", e.getMessage() );
        }
        
        //
        // Error value cannot be arbitrarily close to zero. Must be bounded.
        // 
        try 
        {
            BloomFilter_mutated bf = new BloomFilter_mutated( 1, 0.000000001 );
            fail( "should have thrown an exception: error value is not bounded by a precision or tolerance value." );
        }
        catch( RuntimeException e ) 
        {
            assertEquals( "Invalid params for bloom filter", e.getMessage() );
        }

        //
        // Creating an unusable bloom filter (zero bits, zero hashes) should fail.
        // 
        try 
        {
            BloomFilter_mutated bf = new BloomFilter_mutated(199, 1.0);
            fail( "should have thrown an exception: created an unusable bloom filter with zero bits and zero hashes." );
        }
        catch( RuntimeException e ) 
        {
            assertEquals( "Invalid params for bloom filter", e.getMessage() );
        }

        //
        // Giving it an error value that is too big should fail.
        // 
        try 
        {
            BloomFilter_mutated bf = new BloomFilter_mutated(199, 100.0);
            fail( "should have thrown an exception: error value is too big." );
        }
        catch( RuntimeException e ) 
        {
            assertEquals( "Invalid params for bloom filter", e.getMessage() );
        }

        //
        // Adding more data than expected should fail.
        // 
        try 
        {
            byte []data = "add more entries than bytes available".getBytes();
            BloomFilter_mutated bf0 = new BloomFilter_mutated(data, 1, 0.1); 
            fail( "should have thrown an exception: too much data." );
        }
        catch( RuntimeException e ) 
        {
            assertEquals( "Expected 1 bytes, got 37", e.getMessage() );
        }

    }
    

    public void testValuesFromPublicAPI()
    {
        BloomFilter_mutated bf = null;
        assertEquals(0.000000001, bf.errorPrecision);
        
        bf = new BloomFilter_mutated(1, 0.01);
        assertEquals(2, bf.bytes);
        assertEquals(1, bf.entries);
        assertEquals(0.01, bf.error);
        assertEquals(9, bf.bits);
        assertEquals(7, bf.hashes);
        //
        bf = new BloomFilter_mutated(1, 0.1);
        assertEquals(1, bf.bytes);
        assertEquals(1, bf.entries);
        assertEquals(0.1, bf.error);
        assertEquals(4, bf.bits);
        assertEquals(4, bf.hashes);        
        //
        bf = new BloomFilter_mutated(8, 0.000001);
        assertEquals(29, bf.bytes);
        assertEquals(8, bf.entries);
        assertEquals(0.000001, bf.error);
        assertEquals(230, bf.bits);
        assertEquals(20, bf.hashes);        
    }

    
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
