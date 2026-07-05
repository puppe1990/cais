// pkg/cais/js/logic/chat.mjs
function shouldApplyChatPoll(pollURL, chatEnabled) {
  if (!pollURL || typeof pollURL !== "string" || pollURL.trim() === "") {
    return false;
  }
  if (!chatEnabled) {
    return false;
  }
  return true;
}
function isWithinFallbackWindow(ms) {
  if (typeof ms !== "number" || ms <= 0) return false;
  return ms < 3e4;
}
export {
  isWithinFallbackWindow,
  shouldApplyChatPoll
};
