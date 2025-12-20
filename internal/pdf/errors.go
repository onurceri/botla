package pdf

import "fmt"

type PDFError struct {
	Op  string
	Err error
}

func (e *PDFError) Error() string {
	return fmt.Sprintf("pdf %s: %v", e.Op, e.Err)
}

func (e *PDFError) Unwrap() error {
	return e.Err
}
