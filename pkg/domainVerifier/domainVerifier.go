package domainVerifier

import (
	"fmt"
	"net"
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
		if len(record) > 5 && record[:6] == "v=spf1" {
			return nil
		}
	}
	return fmt.Errorf("valid SPF record not found for %s", domain)
}

func VerifyDKIMRecord(domain, selector string) error {
	txtRecords, err := net.LookupTXT(fmt.Sprintf("%s._domainkey.%s", selector, domain))
	if err != nil {
		return fmt.Errorf("failed to lookup DKIM record for %s: %v", domain, err)
	}
	for _, record := range txtRecords {
		if len(record) > 7 && record[:8] == "v=DKIM1" {
			return nil
		}
	}
	return fmt.Errorf("valid DKIM record not found for %s", domain)
}

func VerifyDMARCRecord(domain string) error {
	txtRecords, err := net.LookupTXT("_dmarc." + domain)
	if err != nil {
		return fmt.Errorf("failed to lookup DMARC record for %s: %v", domain, err)
	}
	for _, record := range txtRecords {
		if len(record) > 7 && record[:8] == "v=DMARC1" {
			return nil
		}
	}
	return fmt.Errorf("valid DMARC record not found for %s", domain)
}

func VerifyDomain(domain string) (map[string]bool, error) {
	results := make(map[string]bool)

	if err := VerifyMXRecord(domain); err == nil {
		results["MX"] = true
	} else {
		results["MX"] = false
	}

	if err := VerifySPFRecord(domain); err == nil {
		results["SPF"] = true
	} else {
		results["SPF"] = false
	}

	if err := VerifyDKIMRecord(domain, "default"); err == nil {
		results["DKIM"] = true
	} else {
		results["DKIM"] = false
	}

	if err := VerifyDMARCRecord(domain); err == nil {
		results["DMARC"] = true
	} else {
		results["DMARC"] = false
	}

	return results, nil
}

