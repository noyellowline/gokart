package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/noyellowline/gokart/internal/core/config"
	"github.com/noyellowline/gokart/internal/core/errors"
)

type Proxy struct {
	reverseProxy *httputil.ReverseProxy
	target       *url.URL
}

func New(cfg *config.ProxyConfig) (*Proxy, error) {
	target, err := url.Parse(cfg.Target)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		MaxIdleConns:    cfg.MaxIdleConns,
		IdleConnTimeout: cfg.IdleConnTimeout,
	}

	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			slog.Debug("proxying request",
				"method", req.Method,
				"path", req.URL.Path,
				"target", target.String())
		},
		Transport:     transport,
		FlushInterval: cfg.FlushInterval,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("proxy error",
				"error", err,
				"method", r.Method,
				"path", r.URL.Path,
				"target", target.String())

			errors.RespondJSON(w, http.StatusBadGateway, "service unavailable")
		},
	}

	slog.Info("proxy initialized", "target", target.String())

	return &Proxy{
		reverseProxy: reverseProxy,
		target:       target,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.reverseProxy.ServeHTTP(w, r)
}
