package main

import "time"

func dateFmt(tt time.Time) string {
	const layout = "2 Jan 2006"
	return tt.Format(layout)
}
