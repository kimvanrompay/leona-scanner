package handler

import "strings"

// isBusinessEmail validates that the email is from a business domain
// Rejects consumer email providers (Gmail, Outlook, etc.)
func isBusinessEmail(email string) bool {
	// Lijst met verboden consumer domeinen
	forbidden := []string{
		"gmail.com",
		"googlemail.com",
		"outlook.com",
		"hotmail.com",
		"live.com",
		"live.nl",
		"live.be",
		"icloud.com",
		"me.com",
		"yahoo.com",
		"yahoo.be",
		"yahoo.nl",
		"telenet.be",
		"skynet.be",
		"proximus.be",
		"pandora.be",
		"scarlet.be",
		"aol.com",
		"gmx.com",
		"gmx.net",
		"mail.com",
		"zoho.com",
		"protonmail.com",
		"tutanota.com",
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])
	for _, d := range forbidden {
		if domain == d {
			return false
		}
	}
	return true
}

// getBusinessEmailError returns a user-friendly error message in Dutch
func getBusinessEmailError() string {
	return "Gebruik a.u.b. een zakelijk e-mailadres voor de CRA-audit. Privé-emailadressen (Gmail, Outlook, etc.) worden niet geaccepteerd."
}
