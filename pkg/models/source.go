package models

type SourceType string

const (
	SourceTypeHttp = "http"
)

var SourceTypes = []SourceType{
	SourceTypeHttp,
}

type VerificationType string

const (
	VerificationTypeHMAC VerificationType = "hmac"
)

var VerificationTypes = []VerificationType{
	VerificationTypeHMAC,
}

type HMACAlgorithm string

const (
	HMACAlgorithmSHA256 HMACAlgorithm = "sha256"
)

var HMACAlgorithms = []HMACAlgorithm{
	HMACAlgorithmSHA256,
}

type Source struct {
	ID           string        `yaml:"id"`
	Slug         string        `yaml:"slug"`
	Type         SourceType    `yaml:"type"`
	Verification *Verification `yaml:"verification"`
}
