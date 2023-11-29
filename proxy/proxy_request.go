package proxy

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/telegraf-tag-auth-proxy/middleware"
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
)

// Inspiration taken from here: https://gist.github.com/yowu/f7dc34bd4736a65ff28d
var upstreamTimeout = env.Duration("PROXY_UPSTREAM_TIMEOUT", 0) // TODO: move this stuff to some struct for easier testing

var dropHeaders = []string{
	// Hop-by-hop headers. These are removed when sent to the backend.
	// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",

	// Also drop authentication headers, this might be overwritten anyway
	"Authorization",
}

// copyHeaders copies all headers except the ones defined in dropHeaders
func copyHeaders(dst, src http.Header) {
	for header, values := range src {
		if slices.Contains(dropHeaders, header) {
			// skip copying hop-by-hop headers
			continue
		}
		for _, val := range values {
			dst.Add(header, val)
		}
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func ProxyRequest(body []byte, ctx *gin.Context, upstreamURL string) {
	cl := http.Client{
		Timeout: upstreamTimeout,
	}

	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequestWithContext(ctx, ctx.Request.Method, upstreamURL, bodyReader)
	copyHeaders(req.Header, ctx.Request.Header)

	var clientIP string
	clientIP, _, err = net.SplitHostPort(ctx.Request.RemoteAddr)
	if err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	var resp *http.Response
	resp, err = cl.Do(req)
	if err != nil {
		middleware.GetLoggerFromCtx(ctx).Error("Error while making request to upstream", zap.Error(err))
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		e := Body.Close()
		if e != nil {
			middleware.GetLoggerFromCtx(ctx).Error("Failed to close upstream body", zap.Error(e))
		}
	}(resp.Body)

	copyHeaders(ctx.Writer.Header(), resp.Header)
	ctx.Status(resp.StatusCode)

	var copied int64
	copied, err = io.Copy(ctx.Writer, resp.Body)
	if err != nil {
		middleware.GetLoggerFromCtx(ctx).Error("Failed to copy body", zap.Int64("copied", copied), zap.Error(err))
	}
	middleware.GetLoggerFromCtx(ctx).Debug("Proxy complete", zap.Int64("upstreamResponseBodyLen", copied), zap.Int64("donwstreamRequestBodyLen", req.ContentLength), zap.Int("upstreamStatusCode", resp.StatusCode))
}
