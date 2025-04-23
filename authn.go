package main

type Authenticator struct {
	db *PasswordDatabase
}

func (authn *Authenticator) Authenticate(service, username, password string) error {
	return authn.db.Authenticate(username, []byte(password))
}
