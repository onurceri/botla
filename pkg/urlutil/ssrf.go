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

// NewSSRFValidator creates a new SSRF validator with the specified configuration.
// Set allowPrivate to true to allow private IPs (useful for testing).
func NewSSRFValidator(allowPrivate bool) *SSRFValidator {
	return &SSRFValidator{allowPrivate: allowPrivate}
}

// SetAllowPrivate allows enabling/disabling private IP checks at runtime.
// This is primarily used for testing.
func (v *SSRFValidator) SetAllowPrivate(allow bool) {
	v.allowPrivate = allow
}

// BlockedSchemes are URL schemes that should never be allowed.
// Note: These are documented here for reference; validation explicitly
// requires http/https schemes, so these are implicitly blocked.
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
	"10.0.0.0/8",     // Private Class A
	"172.16.0.0/12",  // Private Class B
	"192.168.0.0/16", // Private Class C
	"127.0.0.0/8",    // Loopback
	"169.254.0.0/16", // Link-local (includes cloud metadata)
	"::1/128",        // IPv6 loopback
	"fc00::/7",       // IPv6 private
	"fe80::/10",      // IPv6 link-local
	"100.64.0.0/10",  // Carrier-grade NAT
	"0.0.0.0/8",      // Current network
}

var parsedBlockedRanges []*net.IPNet

func init() {
	for _, cidr := range BlockedIPRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid CIDR %q in BlockedIPRanges: %v", cidr, err))
		}
		parsedBlockedRanges = append(parsedBlockedRanges, network)
	}
}

// ValidateURL checks if a URL is safe from SSRF attacks.
// Returns an error if the URL is blocked or invalid.
func (v *SSRFValidator) ValidateURL(rawURL string) error {
	_, err := v.validateAndResolve(rawURL)
	return err
}

// ValidateURLStrict performs validation and also returns resolved IPs.
// Use this for logging what IP was actually contacted.
// This avoids double DNS lookups compared to calling ValidateURL + LookupIP separately.
func (v *SSRFValidator) ValidateURLStrict(rawURL string) ([]net.IP, error) {
	return v.validateAndResolve(rawURL)
}

// validateAndResolve is the core validation logic that parses, validates,
// and optionally resolves DNS for the given URL.
func (v *SSRFValidator) validateAndResolve(rawURL string) ([]net.IP, error) {
	// Parse URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme - only http and https are allowed
	scheme := strings.ToLower(parsed.Scheme)
	if scheme == "" {
		return nil, fmt.Errorf("missing URL scheme")
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("blocked URL scheme: %s", scheme)
	}

	// Check hostname
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return nil, fmt.Errorf("missing hostname")
	}

	for _, blocked := range BlockedHosts {
		if host == blocked {
			if v.allowPrivate && (host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "[::1]") {
				continue
			}
			return nil, fmt.Errorf("blocked hostname: %s", host)
		}
	}

	// Check for IP address directly in URL
	if ip := net.ParseIP(host); ip != nil {
		if ipErr := v.validateIP(ip); ipErr != nil {
			return nil, ipErr
		}
		// Return this single IP for direct IP URLs
		return []net.IP{ip}, nil
	}

	// Resolve hostname and validate IPs
	if v.allowPrivate {
		// In test mode, skip DNS resolution
		return nil, nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname: %w", err)
	}

	for _, ip := range ips {
		if err := v.validateIP(ip); err != nil {
			return nil, err
		}
	}

	return ips, nil
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
