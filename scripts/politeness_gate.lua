local function politness_gate(key)
    local reply = redis.call("HGETALL", key)
    if not reply or #reply == 0 then
        redis.log(redis.LOG_WARNING, "Key not found or empty hash: " .. key)
        return { nil, false }
    end

    local data = {}
    for i = 1, #reply, 2 do
        data[reply[i]] = reply[i + 1]
    end

    local tokens_num = tonumber(data["tokens_num"])
    local max_tokens_num = tonumber(data["max_tokens_num"])
    local refill_time = tonumber(data["refill_time"])
    local delay = tonumber(data["delay"])
    local current_time = tonumber(redis.call("TIME")[1])
    if not tokens_num or not max_tokens_num or not refill_time or not delay then
        redis.log(redis.LOG_WARNING, "Invalid or missing hash fields: " .. key)
        return { nil, false }
    end

    if refill_time > current_time then
        return { nil, false }
    end
    if tokens_num >= max_tokens_num and max_tokens_num ~= 0 then
        redis.log(redis.LOG_WARNING, "TOKENS ARE TOO MUCH: ", tokens_num, max_tokens_num)
        tokens_num = 0
        refill_time = current_time + delay
        redis.call("HSET", key, "tokens_num", tokens_num)
        redis.call("HSET", key, "refill_time", refill_time)
        return { nil, false }
    end
    if refill_time < current_time then
        tokens_num = tokens_num + 1
        redis.call("HSET", key, "tokens_num", tokens_num)
        return { data["rules"], true }
    end

    return nil
end

return politness_gate(KEYS[1])
