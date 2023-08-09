package models

type Verification struct {
	VerificationType     VerificationType `yaml:"verificationType"`
	HMACAlgorithm        *HMACAlgorithm   `yaml:"hmacAlgorithm"`
	SignatureHeader      string           `yaml:"signatureHeader"`
	SignaturePrefix      string           `yaml:"signaturePrefix"`
	CurrentSecretEnvVar  string           `yaml:"currentSecretEnvVar"`
	PreviousSecretEnvVar string           `yaml:"previousSecretEnvVar"`
}
