package virtualmin

import (
	"fmt"
	"os/exec"
	"strings"
)

var Log func(string, ...interface{})

func emailDomainName(email string) (string, error) {
	trimmedEmail := strings.TrimSpace(email)
	lastPos := strings.LastIndex(trimmedEmail, "@")
	if lastPos < 0 {
		return "", fmt.Errorf("@ missing, not valid email: " + email)
	}

	domain := trimmedEmail[lastPos+1:]
	Log("domain %s, email: %s", domain, email)

	return domain, nil
}

func SuspendEmail(email string) (error) {
	domain, err := emailDomainName(email)
	if (err != nil) {
		return err
	}
	username := strings.Split(email, "@")[0]
	Log("disable email: %s, domain: %s, username: %s", email, domain, username)
	// "virtualmin modify-user --domain s12.mercstudio.com --user james --disable-email"
	cmd := exec.Command("/usr/sbin/virtualmin", "modify-user", "--domain", domain, "--user", username, "--disable")
	output, err := cmd.CombinedOutput()
	Log("output: %s", string(output))
	if err != nil {
		Log("command failed: %v. Output: %s", err, string(output))
		return fmt.Errorf("virtualmin command failed to suspend email %s: %w. Output: %s", email, err, string(output))
	}
	return nil
}

func EnableEmail(email string) (error) {
	domain, err := emailDomainName(email)
	if (err != nil) {
		return err
	}
	username := strings.Split(email, "@")[0]
	Log("enable email: %s, domain: %s, username: %s", email, domain, username) // Corrected log message
	// "virtualmin modify-user --domain s12.mercstudio.com --user james --enable-email" // Corrected comment
	cmd := exec.Command("/usr/sbin/virtualmin", "modify-user", "--domain", domain, "--user", username, "--enable")
	output, err := cmd.CombinedOutput()
	Log("output: %s", string(output))
	if err != nil {
		Log("command failed: %v. Output: %s", err, string(output))
		return fmt.Errorf("virtualmin command failed to enable email %s: %w. Output: %s", email, err, string(output))
	}
	return nil
}
