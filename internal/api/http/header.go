package http

import (
	"context"
	"net"
	"net/http"
	"strings"
)

const (
	Authorization   = "authorization"
	Accept          = "accept"
	AcceptLanguage  = "accept-language"
	CacheControl    = "cache-control"
	ContentType     = "content-type"
	Expires         = "expires"
	Location        = "location"
	Origin          = "origin"
	Pragma          = "pragma"
	UserAgentHeader = "user-agent"
	ForwardedFor    = "x-forwarded-for"

	ContentSecurityPolicy   = "content-security-policy"
	XXSSProtection          = "x-xss-protection"
	StrictTransportSecurity = "strict-transport-security"
	XFrameOptions           = "x-frame-options"
	XContentTypeOptions     = "x-content-type-options"
	ReferrerPolicy          = "referrer-policy"
	FeaturePolicy           = "feature-policy"

	ZitadelOrgID = "x-zitadel-orgid"
)

type key int

var (
	httpHeaders key
	remoteAddr  key
)

func CopyHeadersToContext(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), httpHeaders, r.Header)
		ctx = context.WithValue(ctx, remoteAddr, r.RemoteAddr)
		r = r.WithContext(ctx)
		h(w, r)
	}
}

func HeadersFromCtx(ctx context.Context) (http.Header, bool) {
	headers, ok := ctx.Value(httpHeaders).(http.Header)
	return headers, ok
}

func RemoteIPFromCtx(ctx context.Context) string {
	ctxHeaders, ok := HeadersFromCtx(ctx)
	if !ok {
		return RemoteAddrFromCtx(ctx)
	}
	forwarded, ok := GetForwardedFor(ctxHeaders)
	if ok {
		return forwarded
	}
	return RemoteAddrFromCtx(ctx)
}

func RemoteIPFromRequest(r *http.Request) net.IP {
	return net.ParseIP(RemoteIPStringFromRequest(r))
}

func RemoteIPStringFromRequest(r *http.Request) string {
	ip, ok := GetForwardedFor(r.Header)
	if ok {
		return ip
	}
	return r.RemoteAddr
}

func GetForwardedFor(headers http.Header) (string, bool) {
	forwarded, ok := headers[ForwardedFor]
	if ok {
		ip := strings.Split(forwarded[0], ", ")[0]
		if ip != "" {
			return ip, true
		}
	}
	return "", false
}

func RemoteAddrFromCtx(ctx context.Context) string {
	ctxRemoteAddr, _ := ctx.Value(remoteAddr).(string)
	return ctxRemoteAddr
}
