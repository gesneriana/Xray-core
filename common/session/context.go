package session

import (
	"context"
	"github.com/xtls/xray-core/common/net"
	"sync"

	"github.com/xtls/xray-core/features/routing"
)

type sessionKey int

const (
	idSessionKey sessionKey = iota
	inboundSessionKey
	outboundSessionKey
	contentSessionKey
	muxPreferedSessionKey
	sockoptSessionKey
	trackedConnectionErrorKey
	dispatcherKey
)

var once sync.Once
var outboundIp net.IP

// ContextWithID returns a new context with the given ID.
func ContextWithID(ctx context.Context, id ID) context.Context {
	return context.WithValue(ctx, idSessionKey, id)
}

// IDFromContext returns ID in this context, or 0 if not contained.
func IDFromContext(ctx context.Context) ID {
	if id, ok := ctx.Value(idSessionKey).(ID); ok {
		return id
	}
	return 0
}

func ContextWithInbound(ctx context.Context, inbound *Inbound) context.Context {
	return context.WithValue(ctx, inboundSessionKey, inbound)
}

func InboundFromContext(ctx context.Context) *Inbound {
	if inbound, ok := ctx.Value(inboundSessionKey).(*Inbound); ok {
		return inbound
	}
	return nil
}

func ContextWithOutbound(ctx context.Context, outbound *Outbound) context.Context {
	return context.WithValue(ctx, outboundSessionKey, outbound)
}

func OutboundFromContext(ctx context.Context) *Outbound {
	if outbound, ok := ctx.Value(outboundSessionKey).(*Outbound); ok {
		if outboundIp != nil && len(outboundIp) > 0 {
			outbound.Gateway = net.IPAddress(outboundIp)
		}
		return outbound
	}
	return nil
}

func ContextWithContent(ctx context.Context, content *Content) context.Context {
	return context.WithValue(ctx, contentSessionKey, content)
}

func ContentFromContext(ctx context.Context) *Content {
	if content, ok := ctx.Value(contentSessionKey).(*Content); ok {
		return content
	}
	return nil
}

// ContextWithMuxPrefered returns a new context with the given bool
func ContextWithMuxPrefered(ctx context.Context, forced bool) context.Context {
	return context.WithValue(ctx, muxPreferedSessionKey, forced)
}

// MuxPreferedFromContext returns value in this context, or false if not contained.
func MuxPreferedFromContext(ctx context.Context) bool {
	if val, ok := ctx.Value(muxPreferedSessionKey).(bool); ok {
		return val
	}
	return false
}

// ContextWithSockopt returns a new context with Socket configs included
func ContextWithSockopt(ctx context.Context, s *Sockopt) context.Context {
	return context.WithValue(ctx, sockoptSessionKey, s)
}

// SockoptFromContext returns Socket configs in this context, or nil if not contained.
func SockoptFromContext(ctx context.Context) *Sockopt {
	if sockopt, ok := ctx.Value(sockoptSessionKey).(*Sockopt); ok {
		return sockopt
	}
	return nil
}

func GetForcedOutboundTagFromContext(ctx context.Context) string {
	if ContentFromContext(ctx) == nil {
		return ""
	}
	return ContentFromContext(ctx).Attribute("forcedOutboundTag")
}

func SetForcedOutboundTagToContext(ctx context.Context, tag string) context.Context {
	if contentFromContext := ContentFromContext(ctx); contentFromContext == nil {
		ctx = ContextWithContent(ctx, &Content{})
	}
	ContentFromContext(ctx).SetAttribute("forcedOutboundTag", tag)
	return ctx
}

type TrackedRequestErrorFeedback interface {
	SubmitError(err error)
}

func SubmitOutboundErrorToOriginator(ctx context.Context, err error) {
	if errorTracker := ctx.Value(trackedConnectionErrorKey); errorTracker != nil {
		errorTracker := errorTracker.(TrackedRequestErrorFeedback)
		errorTracker.SubmitError(err)
	}
}

func TrackedConnectionError(ctx context.Context, tracker TrackedRequestErrorFeedback) context.Context {
	return context.WithValue(ctx, trackedConnectionErrorKey, tracker)
}

func ContextWithDispatcher(ctx context.Context, dispatcher routing.Dispatcher) context.Context {
	return context.WithValue(ctx, dispatcherKey, dispatcher)
}

func DispatcherFromContext(ctx context.Context) routing.Dispatcher {
	if dispatcher, ok := ctx.Value(dispatcherKey).(routing.Dispatcher); ok {
		return dispatcher
	}
	return nil
}

// SetOutboundIp Solve the traffic infinite loop of local loopback, Can only be set once, otherwise restart the process
func SetOutboundIp(ip net.IP) {
	once.Do(func() {
		outboundIp = ip
	})
}
