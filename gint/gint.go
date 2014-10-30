package gint

func IntMust(i int, err error) int {
	if err != nil {
		panic(err)
	}
	return i
}

func IntIgnore(i int, err error) int {
	return i
}
