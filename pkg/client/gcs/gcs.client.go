package gcs

type Client interface {
	Upload([]byte, string) error
	GetSignedUrl(string) (string, error)
}
