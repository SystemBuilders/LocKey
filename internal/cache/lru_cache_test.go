package cache

import "testing"

func Test_LRUCache(t *testing.T) {
	lruCache := NewLRUCache(5)

	one := NewSimpleKey("1")
	two := NewSimpleKey("2")
	three := NewSimpleKey("3")
	four := NewSimpleKey("4")
	one_1 := NewSimpleKey("1")
	five := NewSimpleKey("5")
	six := NewSimpleKey("6")

	err := lruCache.PutElement(one)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(two)
	if err != nil {

	}
	lruCache.PrintCache()

	err = lruCache.GetElement(one)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(three)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(two)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.PutElement(four)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(one)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(three)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.GetElement(four)
	if err != nil {
		t.Fatal(err)
	}
	lruCache.PrintCache()

	err = lruCache.RemoveElement(one_1)
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
