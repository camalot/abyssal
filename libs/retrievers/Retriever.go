package retrievers

type Retriever interface {
	OutOfDateVersion(pkg interface{}) (bool, string, error)
}