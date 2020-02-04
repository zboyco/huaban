package model

type NilWriter struct {
}

func (w NilWriter) Write(b []byte) (n int, err error) {
	return 0, nil
}
