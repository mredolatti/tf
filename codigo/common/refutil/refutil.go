package refutil

import "time"

// StrRef allocates a string container equal to s on the heap and returns it
func StrRef(s string) *string { return &s }

// TimeRef allocates a time struct on the heap equal to t and returns it
func TimeRef(t time.Time) *time.Time { return &t }
