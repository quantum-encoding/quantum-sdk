# Changelog

## v0.8.0

The reasoning state a tool loop has to hand back, and the cache key that keeps a
conversation on one shard.

### Added
- `ChatRequest.PromptCacheKey` — any stable string the client keeps per
  conversation. The gateway hashes it with the caller's identity and forwards it
  as OpenAI/xAI `prompt_cache_key` (or `x-grok-conv-id` on the xAI
  chat-completions lane), so every turn of one conversation lands on the same
  warm provider cache shard. Empty = derived from the caller's identity alone,
  which puts all of that user's conversations on one shard. Generate one per
  conversation object and reuse it on every turn.

  `/qai/v1/chat` only: the session endpoint derives its key from the session ID
  and ignores a client-supplied one.

- `ContentBlock.Reasoning` (`json.RawMessage`) and `ContentBlock.MintedBy`, on
  blocks of the new type `"reasoning"`. This is the provider's own reasoning
  item, verbatim and opaque. It arrives interleaved with the `tool_use` blocks
  and must be echoed back unchanged, **in the position it arrived in**, on the
  next turn's assistant message: its place among the tool calls is how the
  provider learns where the reasoning sat, and replaying it behind the call it
  reasoned about is a different conversation the provider rejects. Dropping it
  re-bills the reasoning tokens on every round of a tool loop.

  `MintedBy` names the model that produced the block; reasoning state is bound
  to its model and is never replayed to a different one.

  Distinct from the existing `"thinking"` block, which is the human-readable
  summary of the same turn. One is for the reader, one is for the wire.

- `StreamEvent.ThoughtSignature`, carried by the new `thought_signature` SSE
  event the gateway sends just before `done` on a Gemini 3 stream that ended in
  text, and by the atomic `tool_use` event — which a streaming tool loop
  previously had no way to read.

### Changed
- `ContentBlock.ThoughtSignature` is documented on **text** blocks as well as
  `tool_use` blocks: Gemini 3 signs a turn that ends in text. The field already
  accepted it — this states the contract, it is not a shape change.
- `ProviderOptions` is documented as an open map, one JSON object per provider
  key. Newly documented: `provider_options.openai.reasoning_summary`
  (`auto` | `concise` | `detailed` | `none`), `.reasoning_mode`
  (`standard` | `pro`), `.verbosity` (`low` | `medium` | `high`),
  `.text_format` (`text` | `json_object`), and
  `provider_options.xai.native_files` (bool). The type was already
  `map[string]any`, so unnamed keys always rode through.

Additive against an older gateway: the new fields are simply absent.

### Also released in this tag

Work merged after v0.7.0 that had not yet been tagged:

- `chat: the audio share of the input` (a661f84)
- `image: the receipt was on the wire and this SDK dropped it` (8aae892)
- `media: the receipt says how much was made, and what it cost in tokens` (c14bbe1)
- `chat: cache_write_tokens, and the cache split on the streaming event` (46db56f)
- `docs: compute rental requires per-account admin approval` (7c060e3)
- `chore: ship the MIT licence text` (2c3bff1)
- HeyGen v3 surface — avatar realtime sessions, sounds search, template
  detail/render, batch videos (5c13070)
- agentRun retarget, typed 402, Idempotency-Key, balance-after, `cached`,
  streaming `reasoning_tokens` (d8b7e4c)
- HTTP timeout raised 60s → 600s so long media generation is not cut (c1aaa6f)

## v0.7.0 and earlier

Released by git tag only; see `git log v0.7.0`.
