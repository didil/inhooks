-- KEYS[1] = bucket key
-- ARGV[1] = capacity (int)
-- ARGV[2] = refill_rate (int)
-- ARGV[3] = refill_interval_ms (int)
-- ARGV[4] = now_ms (int)
-- Returns: remaining_tokens (float)
--
-- Same as token_bucket.lua but does NOT consume a token or write back.
-- Used for observability/metrics only.

local cap      = tonumber(ARGV[1])
local rate     = tonumber(ARGV[2])
local interval = tonumber(ARGV[3])
local now      = tonumber(ARGV[4])

if not cap or cap <= 0 then
  return redis.error_reply("capacity must be > 0")
end
if not rate or rate <= 0 then
  return redis.error_reply("refill_rate must be > 0")
end
if not interval or interval <= 0 then
  return redis.error_reply("refill_interval must be > 0")
end

local data = redis.call("HMGET", KEYS[1], "tokens", "ts")
local tokens = tonumber(data[1])
local ts     = tonumber(data[2])

if tokens == nil or ts == nil then
  return {cap}
end

local elapsed = now - ts
if elapsed > 0 then
  tokens = math.min(cap, tokens + (elapsed / interval) * rate)
end

return {tokens}
