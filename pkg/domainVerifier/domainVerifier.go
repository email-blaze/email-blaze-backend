package domainVerifier

import (
	"fmt"
	"net"
	"strings"
)

func VerifyMXRecord(domain string) error {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return fmt.Errorf("failed to lookup MX record for %s: %v", domain, err)
	}
	if len(mxRecords) == 0 {
		return fmt.Errorf("no MX records found for %s", domain)
	}
	return nil
}

func VerifySPFRecord(domain string) error {
	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		return fmt.Errorf("failed to lookup SPF record for %s: %v", domain, err)
	}
	for _, record := range txtRecords {
		if strings.HasPrefix(strings.ToLower(record), "v=spf1") {
			return nil
		}
	}
	return fmt.Errorf("valid SPF record not found for %s", domain)
}

func VerifyDKIMRecord(domain, selector string) error {
	dkimDomain := fmt.Sprintf("%s._domainkey.%s", selector, domain)
	txtRecords, err := net.LookupTXT(dkimDomain)
	if err != nil {
		return fmt.Errorf("failed to lookup DKIM record for %s: %v", dkimDomain, err)
	}
	for _, record := range txtRecords {
		if strings.HasPrefix(record, "v=DKIM1") {
			return nil
		}
	}
	return fmt.Errorf("valid DKIM record not found for %s", dkimDomain)
}

func VerifyDMARCRecord(domain string) error {
	dmarcDomain := "_dmarc." + domain
	txtRecords, err := net.LookupTXT(dmarcDomain)
	if err != nil {
		return fmt.Errorf("failed to lookup DMARC record for %s: %v", dmarcDomain, err)
	}
	for _, record := range txtRecords {
		if strings.HasPrefix(record, "v=DMARC1") {
			// Check for required tags
			if strings.Contains(record, "p=") {
				return nil
			}
		}
	}
	return fmt.Errorf("valid DMARC record not found for %s", dmarcDomain)
}

func VerifyDomain(domain string) (map[string]string, error) {
	results := make(map[string]string)
	verifiers := map[string]func(string) error{
		"MX":    VerifyMXRecord,
		"SPF":   VerifySPFRecord,
		"DKIM":  func(d string) error { return VerifyDKIMRecord(d, "default") },
		"DMARC": VerifyDMARCRecord,
	}

	for name, verifier := range verifiers {
		if err := verifier(domain); err == nil {
			results[name] = "Valid"
		} else {
			results[name] = fmt.Sprintf("Invalid: %v", err)
		}
	}

	return results, nil
}

