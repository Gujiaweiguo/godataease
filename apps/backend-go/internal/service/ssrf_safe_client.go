package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const marketTemplateMaxResponseBytes int64 = 10 << 20

var marketTemplateHTTPClient = newMarketTemplateHTTPClient()

var marketTemplateURLValidator = validateMarketTemplateURL

func newMarketTemplateHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{Timeout: 5 * time.Second}

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("ssrf: invalid address %q: %w", addr, err)
		}

		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("ssrf: DNS lookup failed for %q: %w", host, err)
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("ssrf: no IPs resolved for %q", host)
		}

		for _, ipAddr := range ips {
			if isBlockedIP(ipAddr.IP) {
				return nil, fmt.Errorf("ssrf: resolved IP %s is blocked", ipAddr.IP.String())
			}
		}

		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}

	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return fmt.Errorf("ssrf: too many redirects")
		}
		return marketTemplateURLValidator(req.URL.String())
	}

	return client
}

func validateMarketTemplateURL(templateURL string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(templateURL))
	if err != nil {
		return fmt.Errorf("templateUrl is not a valid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("templateUrl scheme must be http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("templateUrl must have a host")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, parsed.Hostname())
	if err != nil {
		return fmt.Errorf("templateUrl hostname lookup failed: %w", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("templateUrl hostname %q did not resolve", parsed.Hostname())
	}

	for _, ipAddr := range ips {
		if isBlockedIP(ipAddr.IP) {
			return fmt.Errorf("templateUrl hostname %q resolves to blocked address %s", parsed.Hostname(), ipAddr.IP.String())
		}
	}

	return nil
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsPrivate() ||
		ip.IsUnspecified()
}
