package main

// meta block in format
//
//	meta {
//	  name: api
//	}
func MetaGenerate(vars map[string]string) string {
	return NameBlockMap("meta", vars)
}
