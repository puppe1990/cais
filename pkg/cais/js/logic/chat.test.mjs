import test from 'node:test';
import assert from 'node:assert/strict';

import {
  shouldApplyChatPoll,
  isWithinFallbackWindow,
} from './chat.mjs';

test('shouldApplyChatPoll returns false when no pollURL', () => {
  assert.equal(shouldApplyChatPoll(null, true), false);
  assert.equal(shouldApplyChatPoll('', true), false);
  assert.equal(shouldApplyChatPoll(undefined, true), false);
});

test('shouldApplyChatPoll returns false when chat not enabled', () => {
  assert.equal(shouldApplyChatPoll('/poll', false), false);
});

test('shouldApplyChatPoll returns true when pollURL present and chat enabled', () => {
  assert.equal(shouldApplyChatPoll('/chat/1/messages', true), true);
});

test('isWithinFallbackWindow clamps and handles zero/negative', () => {
  assert.equal(isWithinFallbackWindow(0), false);
  assert.equal(isWithinFallbackWindow(-10), false);
  assert.equal(isWithinFallbackWindow(1500), true);
  assert.equal(isWithinFallbackWindow(20000), true); // within reasonable window
});
