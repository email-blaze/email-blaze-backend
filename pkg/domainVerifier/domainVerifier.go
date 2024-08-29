package domainverifier

import (
	"fmt"
	"net"
)

func VerifyDomain(domain string) error {
	_, err := net.LookupMX(domain)
	if err != nil {
		return fmt.Errorf("failed to verify domain %s: %v", domain, err)
	}
	return nil
}

func VerifyDKIMRecord(domain, selector string) error {
	txtRecords, err := net.LookupTXT(fmt.Sprintf("%s._domainkey.%s", selector, domain))
	if err != nil {
		return fmt.Errorf("failed to lookup DKIM record for %s: %v", domain, err)
	}

	for _, record := range txtRecords {
		if len(record) > 0 && record[:2] == "v=" {
			return nil
		}
	}

	return fmt.Errorf("valid DKIM record not found for %s", domain)
}