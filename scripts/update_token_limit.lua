local function update_token_limit(key)
    -- FIX: Use HMGET to retrieve multiple fields from a hash
    local values = redis.call("HMGET", key, "tokens_num", "max_tokens_num")
    local tokens_num = tonumber(values[1])
    local max_tokens_num = tonumber(values[2])

    if tokens_num == nil or max_tokens_num == nil then
        return redis.error_reply("Missing or invalid fields for key: " .. key)
    end

    if tokens_num == 0 then
        -- This logic seems to be a reset or initialization if tokens_num is 0
        redis.call("HSET", key, "max_tokens_num", 1)
        return true
    end
    redis.call("HSET", key, "max_tokens_num", tokens_num)
    return true
end

return update_token_limit(KEYS[1])
