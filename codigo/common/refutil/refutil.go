package refutil

import "time"

// StrRef allocates a string container equal to s on the heap and returns a reference to it
func StrRef(s string) *string { return &s }

// TimeRef allocates a time struct on the heap equal to t and returns a reference to it
func TimeRef(t time.Time) *time.Time { return &t }

// Int64Ref allocates an int64 on the heap equal to i and returns a reference to it
func Int64Ref(i int64) *int64 { return &i }
