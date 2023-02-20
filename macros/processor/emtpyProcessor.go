package processor

type emptyProcessor struct{}

func (*emptyProcessor) Replace(url string, macroProvider Provider) (string, error) {
	return "", nil
}
