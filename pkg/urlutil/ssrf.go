package urlutil

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// SSRFValidator validates URLs against SSRF attacks
type SSRFValidator struct {
	allowPrivate bool // For testing only
}

// NewSSRFValidator creates a new SSRF validator
func NewSSRFValidator() *SSRFValidator {
	return &SSRFValidator{allowPrivate: false}
}

// BlockedSchemes are URL schemes that should never be allowed
var BlockedSchemes = []string{
	"file",
	"ftp",
	"gopher",
	"data",
	"javascript",
}

// BlockedHosts are hostnames that should never be allowed
var BlockedHosts = []string{
	"localhost",
	"127.0.0.1",
	"0.0.0.0",
	"[::1]",
	"metadata.google.internal",
}

// BlockedIPRanges are CIDR ranges for private/internal networks
var BlockedIPRanges = []string{
	"10.0.0.0/8",      // Private Class A
	"172.16.0.0/12",   // Private Class B
	"192.168.0.0/16",  // Private Class C
	"127.0.0.0/8",     // Loopback
	"169.254.0.0/16",  // Link-local (includes cloud metadata)
	"::1/128",         // IPv6 loopback
	"fc00::/7",        // IPv6 private
	"fe80::/10",       // IPv6 link-local
	"100.64.0.0/10",   // Carrier-grade NAT
	"0.0.0.0/8",       // Current network
}

var parsedBlockedRanges []*net.IPNet

func init() {
	for _, cidr := range BlockedIPRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			parsedBlockedRanges = append(parsedBlockedRanges, network)
		}
	}
}

// ValidateURL checks if a URL is safe from SSRF attacks
func (v *SSRFValidator) ValidateURL(rawURL string) error {
	// Parse URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		if scheme == "" {
			return fmt.Errorf("missing URL scheme")
		}
		return fmt.Errorf("blocked URL scheme: %s", scheme)
	}

	for _, blocked := range BlockedSchemes {
		if scheme == blocked {
			return fmt.Errorf("blocked URL scheme: %s", scheme)
		}
	}

	// Check hostname
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("missing hostname")
	}

	for _, blocked := range BlockedHosts {
		if host == blocked {
			return fmt.Errorf("blocked hostname: %s", host)
		}
	}

	// Check for IP address directly in URL
	if ip := net.ParseIP(host); ip != nil {
		if err := v.validateIP(ip); err != nil {
			return err
		}
	}

	// Resolve hostname and validate IPs
	if !v.allowPrivate {
		ips, err := net.LookupIP(host)
		if err != nil {
			// If lookup fails, we can't validate IPs, but we already validated the host string
			// Some systems might not have DNS reachable in tests, so we might want to handle this.
			// However, for SSRF protection, being strict is better.
			return fmt.Errorf("failed to resolve hostname: %w", err)
		}

		for _, ip := range ips {
			if err := v.validateIP(ip); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateIP checks if an IP is in a blocked range
func (v *SSRFValidator) validateIP(ip net.IP) error {
	if v.allowPrivate {
		return nil
	}

	// Check if loopback
	if ip.IsLoopback() {
		return fmt.Errorf("loopback address not allowed: %s", ip)
	}

	// Check if private
	if ip.IsPrivate() {
		return fmt.Errorf("private address not allowed: %s", ip)
	}

	// Check if link-local
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("link-local address not allowed: %s", ip)
	}

	// Check against blocked ranges
	for _, network := range parsedBlockedRanges {
		if network.Contains(ip) {
			return fmt.Errorf("address in blocked range: %s", ip)
		}
	}

	return nil
}

// ValidateURLStrict performs validation and also returns resolved IPs
// Use this for logging what IP was actually contacted
func (v *SSRFValidator) ValidateURLStrict(rawURL string) ([]net.IP, error) {
	if err := v.ValidateURL(rawURL); err != nil {
		return nil, err
	}

	parsed, _ := url.Parse(rawURL)
	host := parsed.Hostname()

	// Return resolved IPs for logging
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("lookup ip failed: %w", err)
	}

	return ips, nil
}
