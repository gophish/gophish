// Package ratelimit provides a simple token-bucket rate limiting middleware
// which only allows n POST requests every minute. This is meant to be used on
// login handlers or other sensitive transactions which should be throttled to
// prevent abuse.
//
// Tracked clients are stored in a locked map, with a goroutine that runs at a
// configurable interval to clean up stale entries.
//
// Note that there is no enforcement for GET requests. This is an effort to be
// opinionated in order to hit the most common use-cases. For more advanced
// use-cases, you may consider the `github.com/didip/tollbooth` package.
//
// The enforcement mechanism is based on the blog post here:
// https://www.alexedwards.net/blog/how-to-rate-limit-http-requests
package ratelimit
