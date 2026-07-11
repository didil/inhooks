package services

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/models"
	"github.com/didil/inhooks/pkg/testsupport/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type RateLimiterSuite struct {
	suite.Suite
	timeSvc *mockTimeServiceForRL
	logger  *zap.Logger
	limiter RateLimiter
	store   RedisStore
}

func TestRateLimiterSuite(t *testing.T) {
	suite.Run(t, new(RateLimiterSuite))
}

type mockTimeServiceForRL struct {
	mu  sync.Mutex
	now time.Time
}

func (m *mockTimeServiceForRL) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

func (s *RateLimiterSuite) SetupTest() {
	s.timeSvc = &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	s.logger = zap.NewNop()

	// Use a mock store for basic unit tests
	ctrl := gomock.NewController(s.T())
	mockStore := mocks.NewMockRedisStore(ctrl)
	s.store = mockStore

	// For non-mock tests (real Redis), we'll skip those here
	// The mock store can't run real Lua scripts, so we test logic only
	s.limiter = NewTokenBucketLimiter(mockStore, s.timeSvc, s.logger)
}

func (s *RateLimiterSuite) TestAllow_FailOpenOnStoreError() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)
	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("redis is down"))

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       10,
		RefillRate:     5,
		RefillInterval: time.Second,
	}
	key := "rl:test:failopen"

	decision := limiter.Allow(ctx, key, cfg)
	s.True(decision.Allowed, "fail-open: should allow on store error")
}

func (s *RateLimiterSuite) TestAllow_ParseError_FailOpen() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)

	// Return a malformed result (only 1 element instead of 3)
	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]interface{}{int64(1)}, nil)

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       5,
		RefillRate:     2,
		RefillInterval: time.Second,
	}

	decision := limiter.Allow(ctx, "rl:test", cfg)
	s.True(decision.Allowed, "fail-open on parse error")
}

func (s *RateLimiterSuite) TestAllow_Success() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)

	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]interface{}{int64(1), int64(7), int64(0)}, nil)

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       10,
		RefillRate:     5,
		RefillInterval: time.Second,
	}

	decision := limiter.Allow(ctx, "rl:test", cfg)
	s.True(decision.Allowed)
	s.Equal(float64(7), decision.Remaining)
	s.Equal(time.Duration(0), decision.RetryAfter)
}

func (s *RateLimiterSuite) TestAllow_Denied() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)

	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]interface{}{int64(0), int64(0), int64(5000)}, nil)

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       5,
		RefillRate:     1,
		RefillInterval: time.Second,
	}

	decision := limiter.Allow(ctx, "rl:test", cfg)
	s.False(decision.Allowed)
	s.Equal(float64(0), decision.Remaining)
	s.Equal(5*time.Second, decision.RetryAfter)
}

func (s *RateLimiterSuite) TestAllow_ConcurrentCalls() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)

	// Simulate a real bucket: allow exactly 10
	var allowed int32
	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx interface{}, script interface{}, keys interface{}, args ...interface{}) ([]interface{}, error) {
			n := atomic.AddInt32(&allowed, 1)
			if n <= 10 {
				return []interface{}{int64(1), int64(10 - n), int64(0)}, nil
			}
			return []interface{}{int64(0), int64(0), int64(1000)}, nil
		}).AnyTimes()

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       10,
		RefillRate:     1,
		RefillInterval: time.Hour,
	}
	key := "rl:test:atomic"

	const goroutines = 50
	var allowedCount int32
	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := make(chan struct{})
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			d := limiter.Allow(ctx, key, cfg)
			if d.Allowed {
				atomic.AddInt32(&allowedCount, 1)
			}
		}()
	}

	close(start)
	wg.Wait()

	s.Equal(int32(10), allowedCount, "exactly capacity calls should be allowed under contention")
}

func (s *RateLimiterSuite) TestParseAllowResult_Int64Remaining() {
	// Simulate Redis returning int64 for remaining (happens when no float math in Lua)
	res := []interface{}{int64(1), int64(9), int64(0)}

	allowed, remaining, retryAfter, err := parseAllowResult(res)
	s.NoError(err)
	s.True(allowed)
	s.Equal(float64(9), remaining)
	s.Equal(time.Duration(0), retryAfter)
}

func (s *RateLimiterSuite) TestParseAllowResult_Float64Remaining() {
	// Simulate Redis returning float64 for remaining (happens when Lua does float math)
	res := []interface{}{int64(1), 7.5, int64(0)}

	allowed, remaining, retryAfter, err := parseAllowResult(res)
	s.NoError(err)
	s.True(allowed)
	s.InDelta(7.5, remaining, 0.01)
	s.Equal(time.Duration(0), retryAfter)
}

func (s *RateLimiterSuite) TestParseAllowResult_WrongLength() {
	res := []interface{}{int64(1)}

	_, _, _, err := parseAllowResult(res)
	s.Error(err)
	s.Contains(err.Error(), "expected 3 values")
}

func (s *RateLimiterSuite) TestPeek_Success() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)

	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]interface{}{int64(5)}, nil)

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       10,
		RefillRate:     5,
		RefillInterval: time.Second,
	}

	tokens, err := limiter.Peek(ctx, "rl:test", cfg)
	s.NoError(err)
	s.Equal(float64(5), tokens)
}

func (s *RateLimiterSuite) TestPeek_EmptyResult() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)

	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]interface{}{}, nil)

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       10,
		RefillRate:     5,
		RefillInterval: time.Second,
	}

	_, err := limiter.Peek(ctx, "rl:test", cfg)
	s.Error(err)
	s.Contains(err.Error(), "empty result")
}

func (s *RateLimiterSuite) TestPeek_StoreError() {
	ctrl := gomock.NewController(s.T())
	store := mocks.NewMockRedisStore(ctrl)

	store.EXPECT().Eval(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("connection refused"))

	timeSvc := &mockTimeServiceForRL{now: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	logger := zap.NewNop()
	limiter := NewTokenBucketLimiter(store, timeSvc, logger)

	ctx := context.Background()
	cfg := &models.RateLimitConfig{
		Capacity:       10,
		RefillRate:     5,
		RefillInterval: time.Second,
	}

	_, err := limiter.Peek(ctx, "rl:test", cfg)
	s.Error(err)
}
