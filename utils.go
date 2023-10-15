package main

func makeCache() MemoryCache {
	return MemoryCache{
		Backend:  make(map[string]*Mesg, 0),
		Maxcount: 0,
	}
}

func makeQuestionCache(maxCount int) *MemoryQuestionCache {
	return &MemoryQuestionCache{Backend: make([]QuestionCacheEntry, 0), Maxcount: maxCount}
}

// Difference between to lists: A - B
func difference[T comparable](a, b []T) (diff []T) {
	m := make(map[T]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}
