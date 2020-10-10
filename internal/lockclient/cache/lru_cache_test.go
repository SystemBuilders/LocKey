package cache

import "testing"

func Test_LRUCache(t *testing.T) {
	lruCache := NewLRUCache(5)

	one := NewSimpleKey("1", "owner1")
	two := NewSimpleKey("2", "owner1")
	three := NewSimpleKey("3", "owner1")
	four := NewSimpleKey("4", "owner1")
	oneOne := NewSimpleKey("1", "owner1")
	five := NewSimpleKey("5", "owner1")
	six := NewSimpleKey("6", "owner1")

	err := lruCache.PutElement(one)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(two)
	if err != nil {

	}
	lruCache.PrintCache()

	_, err = lruCache.GetElement(one)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(three)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	_, err = lruCache.GetElement(two)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(four)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	_, err = lruCache.GetElement(one)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	_, err = lruCache.GetElement(three)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	_, err = lruCache.GetElement(four)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.RemoveElement(oneOne)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(five)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	// LRU Cache becomes full, so tail element
	// must be deleted on insertion of six
	err = lruCache.PutElement(six)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

}
