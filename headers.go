package main

// heades block in format
//
//	headers {
//	  User-Agent: {{ua}}
//	}
func HeadersGenerate(heads map[string]string) string {
	return NameBlockMap("headers", heads)
}
