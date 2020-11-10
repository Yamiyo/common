package network

import "time"

type OptionFunc func(*Option)

func defaultOption() *Option {
	return &Option{
		timeout:               30 * time.Second,
		keepAlive:             30 * time.Second,
		tlsHandshakeTimeout:   10 * time.Second,
		expectContinueTimeout: 1 * time.Second,
	}
}

type Option struct {
	timeout               time.Duration
	keepAlive             time.Duration
	tlsHandshakeTimeout   time.Duration
	expectContinueTimeout time.Duration
}

func WithTimeout(duration time.Duration) OptionFunc {
	return OptionFunc(func(opt *Option) {
		opt.timeout = duration
	})
}

func WithKeepAlive(duration time.Duration) OptionFunc {
	return OptionFunc(func(opt *Option) {
		opt.keepAlive = duration
	})
}

func WithTLSHandshakeTimeout(duration time.Duration) OptionFunc {
	return OptionFunc(func(opt *Option) {
		opt.tlsHandshakeTimeout = duration
	})
}

func WithExpectContinueTimeout(duration time.Duration) OptionFunc {
	return OptionFunc(func(opt *Option) {
		opt.expectContinueTimeout = duration
	})
}
