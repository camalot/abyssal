package retrievers

type Retriever interface {
	CheckVersionOutOfDate(pkg interface{}) (needsUpdate bool, current string, expected string, error error)
}