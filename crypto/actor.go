package crypto

type Actor struct {
	Name          string
	SigningKey    *Key
	EncryptionKey *Key
}

func MakeActor(name string) (*Actor, error) {

	signingKey, err := GenerateKey()

	if err != nil {
		return nil, err
	}

	encryptionKey, err := GenerateKey()

	if err != nil {
		return nil, err
	}

	signingSettingsKey, err := AsSettingsKey(signingKey, "signing", "ecdsa")

	if err != nil {
		return nil, err
	}

	encryptionSettingsKey, err := AsSettingsKey(encryptionKey, "encryption", "ecdh")

	if err != nil {
		return nil, err
	}

	return &Actor{
		Name:          name,
		EncryptionKey: encryptionSettingsKey,
		SigningKey:    signingSettingsKey,
	}, nil

}
