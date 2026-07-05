/**
 * Pure, testable chat logic.
 * These will be bundled via esbuild into the IIFE assets (cais-chat.js etc).
 * See #99.
 */

export function shouldApplyChatPoll(pollURL, chatEnabled) {
  if (!pollURL || typeof pollURL !== "string" || pollURL.trim() === "") {
    return false;
  }
  if (!chatEnabled) {
    return false;
  }
  return true;
}

/**
 * Simple helper used by fallback scheduling.
 */
export function isWithinFallbackWindow(ms) {
  if (typeof ms !== "number" || ms <= 0) return false;
  // Arbitrary reasonable upper bound for a fallback delay in tests.
  return ms < 30000;
}
