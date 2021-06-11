package id

type IdHelper interface {
	ClaimId() []byte
	DisownId([]byte)
}
