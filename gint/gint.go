package gint

func IMust(i int64, err error) int64 {
	if err != nil {
		panic(err)
	}
	return i
}

func IIgnore(i int64, err error) int64 {
	return i
}
